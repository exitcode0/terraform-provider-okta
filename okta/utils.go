package okta

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v4/okta"
	v5okta "github.com/okta/okta-sdk-golang/v5/okta"
	"github.com/okta/terraform-provider-okta/sdk"
)

const defaultPaginationLimit int64 = 200

func buildSchema(schemas ...map[string]*schema.Schema) map[string]*schema.Schema {
	r := map[string]*schema.Schema{}
	for _, s := range schemas {
		for key, val := range s {
			r[key] = val
		}
	}
	return r
}

// camel cased strings from okta responses become underscore separated to match
// the terraform configs for state file setting (ie. firstName from okta response becomes first_name)
func camelCaseToUnderscore(s string) string {
	a := []rune(s)

	for i, r := range a {
		if !unicode.IsLower(r) {
			a = append(a, 0)
			a[i] = unicode.ToLower(r)
			copy(a[i+1:], a[i:])
			a[i] = []rune("_")[0]
		}
	}

	s = string(a)

	return s
}

func conditionalRequire(d *schema.ResourceData, propList []string, reason string) error {
	var missing []string

	for _, prop := range propList {
		if _, ok := d.GetOk(prop); !ok {
			missing = append(missing, prop)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing conditionally required fields, reason: '%s', missing fields: %s", reason, strings.Join(missing, ", "))
	}

	return nil
}

// Conditionally validates a slice of strings for required and valid values.
func conditionalValidator(field, typeValue string, require, valid, actual []string) error {
	explanation := fmt.Sprintf("failed conditional validation for field \"%s\" of type \"%s\", it can contain %s", field, typeValue, strings.Join(valid, ", "))

	if len(require) > 0 {
		explanation = fmt.Sprintf("%s and must contain %s", explanation, strings.Join(require, ", "))
	}

	for _, val := range require {
		if !contains(actual, val) {
			return fmt.Errorf("%s, received %s", explanation, strings.Join(actual, ", "))
		}
	}

	for _, val := range actual {
		if !contains(valid, val) {
			return fmt.Errorf("%s, received %s", explanation, strings.Join(actual, ", "))
		}
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func containsInt(codes []int, code int) bool {
	for _, a := range codes {
		if a == code {
			return true
		}
	}
	return false
}

// Ensures at least one element is contained in provided slice. More performant version of contains(..) || contains(..)
func containsOne(s []string, elements ...string) bool {
	for _, a := range s {
		if contains(elements, a) {
			return true
		}
	}
	return false
}

func convertInterfaceToStringSet(purportedSet interface{}) []string {
	return convertInterfaceToStringArr(purportedSet.(*schema.Set).List())
}

func convertInterfaceToStringSetNullable(purportedSet interface{}) []string {
	set, ok := purportedSet.(*schema.Set)
	if ok {
		return convertInterfaceToStringArrNullable(set.List())
	}
	return nil
}

func convertInterfaceToStringArr(purportedList interface{}) []string {
	var arr []string
	rawArr, ok := purportedList.([]interface{})
	if ok {
		arr = convertInterfaceArrToStringArr(rawArr)
	}
	return arr
}

func convertInterfaceArrToStringArr(rawArr []interface{}) []string {
	arr := make([]string, len(rawArr))
	for i, thing := range rawArr {
		if a, ok := thing.(string); ok {
			arr[i] = a
		}
	}
	return arr
}

// Converts interface to string array, if there are no elements it returns nil to conform with optional properties.
func convertInterfaceToStringArrNullable(purportedList interface{}) []string {
	arr := convertInterfaceToStringArr(purportedList)
	if len(arr) < 1 {
		return nil
	}
	return arr
}

func createNestedResourceImporter(fields []string) *schema.ResourceImporter {
	return createCustomNestedResourceImporter(fields, fmt.Sprintf("Expecting the following format %s", strings.Join(fields, "/")))
}

// createCustomNestedResourceImporter Fields making up the ID should be in
// order, for instance, []string{"auth_server_id", "policy_id", "id"} However,
// extra fields can be specified after as well, []string{"auth_server_id",
// "policy_id", "id", "extra"}
func createCustomNestedResourceImporter(fields []string, errMessage string) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
			parts := strings.Split(d.Id(), "/")
			if len(parts) != len(fields) {
				return nil, fmt.Errorf("expected %d import fields %q, got %d fields %q", len(fields), strings.Join(fields, "/"), len(parts), d.Id())
			}
			for i, field := range fields {
				if field == "id" {
					d.SetId(parts[i])
					continue
				}
				var value interface{}
				if i < len(parts) {
					// deal with the import parameter being a boolean "true" / "false"
					if bValue, err := strconv.ParseBool(parts[i]); err == nil {
						value = bValue
					} else {
						value = parts[i]
					}
				}
				//lintignore:R001
				_ = d.Set(field, value)
			}

			return []*schema.ResourceData{d}, nil
		},
	}
}

func convertStringSliceToInterfaceSlice(stringList []string) []interface{} {
	if len(stringList) == 0 {
		return nil
	}
	arr := make([]interface{}, len(stringList))
	for i, str := range stringList {
		arr[i] = str
	}
	return arr
}

func convertStringSliceToSet(stringList []string) *schema.Set {
	arr := make([]interface{}, len(stringList))
	for i, str := range stringList {
		arr[i] = str
	}
	return schema.NewSet(schema.HashString, arr)
}

func convertStringSliceToSetNullable(stringList []string) *schema.Set {
	if len(stringList) == 0 {
		return nil
	}
	arr := make([]interface{}, len(stringList))
	for i, str := range stringList {
		arr[i] = str
	}
	return schema.NewSet(schema.HashString, arr)
}

func createValueDiffSuppression(newValueToIgnore string) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		return new == newValueToIgnore
	}
}

func ensureNotDefault(d *schema.ResourceData, t string) error {
	thing := fmt.Sprintf("Default %s", t)

	if d.Get("name").(string) == thing {
		return fmt.Errorf("%s is immutable", thing)
	}

	return nil
}

func getMapString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if res, ok := v.(string); ok {
			return res
		}
	}
	return ""
}

// boolPtr return bool pointer to b's value
func boolPtr(b bool) (ptr *bool) {
	ptr = &b
	return
}

// boolFromBoolPtr if b is nil returns false, otherwise return boolean value of b
func boolFromBoolPtr(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func stringPtr(s string) (ptr *string) {
	ptr = &s
	return
}

func doesResourceExist(response *sdk.Response, err error) (bool, error) {
	if response == nil {
		return false, err
	}
	// We don't want to consider a 404 an error in some cases and thus the delineation
	if response.StatusCode == 404 {
		return false, nil
	}
	if err != nil {
		return false, responseErr(response, err)
	}

	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return false, responseErr(response, err)
	}
	// some of the API response can be 200 and return an empty object or list meaning nothing was found
	body := string(b)
	if body == "{}" || body == "[]" {
		return false, nil
	}

	return true, nil
}

func doesResourceExistV3(response *okta.APIResponse, err error) (bool, error) {
	if response == nil {
		return false, err
	}
	// We don't want to consider a 404 an error in some cases and thus the delineation
	if response.StatusCode == 404 {
		return false, nil
	}
	if err != nil {
		return false, responseErrV3(response, err)
	}

	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return false, responseErrV3(response, err)
	}
	// some of the API response can be 200 and return an empty object or list meaning nothing was found
	body := string(b)
	if body == "{}" || body == "[]" {
		return false, nil
	}

	return true, nil
}

// Useful shortcut for suppressing errors from Okta's SDK when a resource does not exist. Usually used during deletion
// of nested resources.
func suppressErrorOn404(resp *sdk.Response, err error) error {
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return responseErr(resp, err)
}

// TODO switch to suppressErrorOn404 when migration complete
func v3suppressErrorOn404(resp *okta.APIResponse, err error) error {
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return v3responseErr(resp, err)
}

// TODO switch to suppressErrorOn404 when migration complete
func v5suppressErrorOn404(resp *v5okta.APIResponse, err error) error {
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return v5responseErr(resp, err)
}

// Useful shortcut for suppressing errors from Okta's SDK when a Org does not
// have permission to access a feature.
func suppressErrorOn401(what string, meta interface{}, resp *sdk.Response, err error) error {
	if resp != nil && resp.StatusCode == http.StatusUnauthorized {
		logger(meta).Warn(fmt.Sprintf("Suppressing %q on %q", "401 Unauthorized", what))
		return nil
	}
	return responseErr(resp, err)
}

// Useful shortcut for suppressing errors from Okta's SDK when a Org does not
// have permission to access a feature.
func suppressErrorOn403(what string, meta interface{}, resp *sdk.Response, err error) error {
	if resp != nil && resp.StatusCode == http.StatusForbidden {
		logger(meta).Warn(fmt.Sprintf("Suppressing %q on %q", "403 Forbidden", what))
		return nil
	}
	return responseErr(resp, err)
}

func getOktaClientFromMetadata(meta interface{}) *sdk.Client {
	return meta.(*Config).oktaSDKClientV2
}

func getOktaV3ClientFromMetadata(meta interface{}) *okta.APIClient {
	return meta.(*Config).oktaSDKClientV3
}

func getOktaV5ClientFromMetadata(meta interface{}) *v5okta.APIClient {
	return meta.(*Config).oktaSDKClientV5
}

func getAPISupplementFromMetadata(meta interface{}) *sdk.APISupplement {
	return meta.(*Config).oktaSDKsupplementClient
}

func getRequestExecutor(m interface{}) *sdk.RequestExecutor {
	return getOktaClientFromMetadata(m).GetRequestExecutor()
}

func is404(resp *sdk.Response) bool {
	return resp != nil && resp.StatusCode == http.StatusNotFound
}

func logger(meta interface{}) hclog.Logger {
	return meta.(*Config).logger
}

func normalizeDataJSON(val interface{}) string {
	dataMap := map[string]interface{}{}

	// Ignoring errors since we know it is valid
	_ = json.Unmarshal([]byte(val.(string)), &dataMap)
	ret, _ := json.Marshal(dataMap)

	return string(ret)
}

// Removes nulls from group profile map and returns, since Okta does not render nulls in profile
func normalizeGroupProfile(profile sdk.GroupProfileMap) sdk.GroupProfileMap {
	trimedProfile := make(sdk.GroupProfileMap)
	for k, v := range profile {
		if v != nil {
			trimedProfile[k] = v
		}
	}
	return trimedProfile
}

// Opposite of append
func remove(arr []string, el string) []string {
	var newArr []string

	for _, item := range arr {
		if item != el {
			newArr = append(newArr, item)
		}
	}
	return newArr
}

// appendUnique appends el to arr if el isn't already present in arr
func appendUnique(arr []string, el string) []string {
	found := false
	for _, item := range arr {
		if item == el {
			found = true
			break
		}
	}
	if found {
		return arr
	}
	return append(arr, el)
}

// The best practices states that aggregate types should have error handling (think non-primitive). This will not attempt to set nil values.
func setNonPrimitives(d *schema.ResourceData, valueMap map[string]interface{}) error {
	for k, v := range valueMap {
		if v != nil {
			//lintignore:R001
			if err := d.Set(k, v); err != nil {
				return fmt.Errorf("error setting %s for resource %s: %s", k, d.Id(), err)
			}
		}
	}
	return nil
}

// Okta SDK will (not often) return just `Okta API has returned an error: ""“ when the error is not valid JSON.
// The status should help with debugability. Potentially also could check for an empty error and omit
// it when it occurs and build some more context.
func responseErr(resp *sdk.Response, err error) error {
	if err != nil {
		msg := err.Error()
		if resp != nil {
			msg += fmt.Sprintf(", Status: %s", resp.Status)
		}
		return errors.New(msg)
	}
	return nil
}

func responseErrV3(resp *okta.APIResponse, err error) error {
	if err != nil {
		msg := err.Error()
		if resp != nil {
			msg += fmt.Sprintf(", Status: %s", resp.Status)
		}
		return errors.New(msg)
	}
	return nil
}

// TODO switch to responseErr when migration complete
func v3responseErr(resp *okta.APIResponse, err error) error {
	if err != nil {
		msg := err.Error()
		if resp != nil {
			msg += fmt.Sprintf(", Status: %s", resp.Status)
		}
		return errors.New(msg)
	}
	return nil
}

// TODO switch to responseErr when migration complete
func v5responseErr(resp *v5okta.APIResponse, err error) error {
	if err != nil {
		msg := err.Error()
		if resp != nil {
			msg += fmt.Sprintf(", Status: %s", resp.Status)
		}
		return errors.New(msg)
	}
	return nil
}

func validatePriority(in, out int64) error {
	if in > 0 && in != out {
		return fmt.Errorf("provided priority was not valid, got: %d, API responded with: %d. See schema for attribute details", in, out)
	}
	return nil
}

func buildEnum(ae []interface{}, elemType string) ([]interface{}, error) {
	enum := make([]interface{}, len(ae))
	for i, aeVal := range ae {
		if aeVal == nil {
			switch elemType {
			case "number":
				enum[i] = float64(0)
			case "integer":
				enum[i] = 0
			default:
				enum[i] = ""
			}
			continue
		}

		aeStr, ok := aeVal.(string)
		if !ok {
			return nil, fmt.Errorf("expected %+v value to cast to string", aeVal)
		}
		switch elemType {
		case "number":
			f, err := strconv.ParseFloat(aeStr, 64)
			if err != nil {
				return nil, errInvalidElemFormat
			}
			enum[i] = f
		case "integer":
			f, err := strconv.Atoi(aeStr)
			if err != nil {
				return nil, errInvalidElemFormat
			}
			enum[i] = f
		default:
			enum[i] = aeStr
		}
	}
	return enum, nil
}

// localFileStateFunc - helper for schema.Schema StateFunc checking if a the
// blob of a local file has changed - is not file path dependant.
func localFileStateFunc(val interface{}) string {
	filePath := val.(string)
	if filePath == "" {
		return ""
	}
	return computeFileHash(filePath)
}

// computeFileHash - equivalent to  `shasum -a 256 filepath`
func computeFileHash(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer file.Close()
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// suppressDuringCreateFunc - attribute has changed assume this is a create and
// treat the properties as readers not caring about what would otherwise apear
// to be drift.
func suppressDuringCreateFunc(attribute string) func(k, old, new string, d *schema.ResourceData) bool {
	return func(k, old, new string, d *schema.ResourceData) bool {
		if d.HasChange(attribute) {
			return true
		}
		return old == new
	}
}

// Normalizes to certificate object when it's passed as a raw b64 block instead of a full pem file
func rawCertNormalize(certContents string) (*x509.Certificate, error) {
	certContents = strings.ReplaceAll(strings.TrimSpace(certContents), " ", "")
	certDecoded, err := base64.StdEncoding.DecodeString(certContents)
	if err != nil {
		return nil, fmt.Errorf("failed to decode b64: %s", err)
	}
	cert, err := x509.ParseCertificate(certDecoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode pem certificate: %s", err)
	}

	return cert, nil
}

// Normalizes to certificate object when passed as PEM formatted certificate file
func pemCertNormalize(certContents string) (*x509.Certificate, error) {
	certContents = strings.TrimSpace(certContents)
	cert, rest := pem.Decode([]byte(certContents))
	if cert == nil {
		return nil, fmt.Errorf("failed to decode PEM file, rest: %s", rest)
	}

	parsedCert, err := x509.ParseCertificate(cert.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %s", err)
	}

	return parsedCert, nil
}

func certNormalize(certContents string) (*x509.Certificate, error) {
	certDecoded, err := pemCertNormalize(certContents)
	if err == nil {
		return certDecoded, nil
	}
	certDecoded, err = rawCertNormalize(certContents)
	if err != nil {
		return nil, err
	}
	return certDecoded, nil
}

// noChangeInObjectFromUnmarshaledJSON Intended for use by a DiffSuppressFunc,
// returns true if old and new JSONs are equivalent object representations ...
// It is true, there is no change!  Edge chase if newJSON is blank, will also
// return true which cover the new resource case.
func noChangeInObjectFromUnmarshaledJSON(k, oldJSON, newJSON string, d *schema.ResourceData) bool {
	if newJSON == "" {
		return true
	}
	var oldObj map[string]any
	var newObj map[string]any
	if err := json.Unmarshal([]byte(oldJSON), &oldObj); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(newJSON), &newObj); err != nil {
		return false
	}

	return reflect.DeepEqual(oldObj, newObj)
}

func Intersection(old []string, new []string) (intersection []string, exclusiveOld []string, exclusiveNew []string) {
	intersection = make([]string, 0)
	exclusiveOld = make([]string, 0)
	exclusiveNew = make([]string, 0)
	oldElementMap := make(map[string]bool)
	newElementMap := make(map[string]bool)
	for _, o := range old {
		oldElementMap[o] = true
	}
	for _, n := range new {
		newElementMap[n] = true
	}
	for _, n := range new {
		if oldElementMap[n] {
			intersection = append(intersection, n)
		} else {
			exclusiveNew = append(exclusiveNew, n)
		}
	}
	for _, o := range old {
		if !newElementMap[o] {
			exclusiveOld = append(exclusiveOld, o)
		}
	}
	return
}
