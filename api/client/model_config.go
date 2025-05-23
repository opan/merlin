/*
Merlin

API Guide for accessing Merlin's model management, deployment, and serving functionalities

API version: 0.14.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package client

import (
	"encoding/json"
)

// checks if the Config type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Config{}

// Config struct for Config
type Config struct {
	JobConfig                   *PredictionJobConfig          `json:"job_config,omitempty"`
	ImageRef                    *string                       `json:"image_ref,omitempty"`
	ServiceAccountName          *string                       `json:"service_account_name,omitempty"`
	ResourceRequest             *PredictionJobResourceRequest `json:"resource_request,omitempty"`
	ImageBuilderResourceRequest *ResourceRequest              `json:"image_builder_resource_request,omitempty"`
	EnvVars                     []EnvVar                      `json:"env_vars,omitempty"`
	Secrets                     []MountedMLPSecret            `json:"secrets,omitempty"`
}

// NewConfig instantiates a new Config object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewConfig() *Config {
	this := Config{}
	return &this
}

// NewConfigWithDefaults instantiates a new Config object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewConfigWithDefaults() *Config {
	this := Config{}
	return &this
}

// GetJobConfig returns the JobConfig field value if set, zero value otherwise.
func (o *Config) GetJobConfig() PredictionJobConfig {
	if o == nil || IsNil(o.JobConfig) {
		var ret PredictionJobConfig
		return ret
	}
	return *o.JobConfig
}

// GetJobConfigOk returns a tuple with the JobConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Config) GetJobConfigOk() (*PredictionJobConfig, bool) {
	if o == nil || IsNil(o.JobConfig) {
		return nil, false
	}
	return o.JobConfig, true
}

// HasJobConfig returns a boolean if a field has been set.
func (o *Config) HasJobConfig() bool {
	if o != nil && !IsNil(o.JobConfig) {
		return true
	}

	return false
}

// SetJobConfig gets a reference to the given PredictionJobConfig and assigns it to the JobConfig field.
func (o *Config) SetJobConfig(v PredictionJobConfig) {
	o.JobConfig = &v
}

// GetImageRef returns the ImageRef field value if set, zero value otherwise.
func (o *Config) GetImageRef() string {
	if o == nil || IsNil(o.ImageRef) {
		var ret string
		return ret
	}
	return *o.ImageRef
}

// GetImageRefOk returns a tuple with the ImageRef field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Config) GetImageRefOk() (*string, bool) {
	if o == nil || IsNil(o.ImageRef) {
		return nil, false
	}
	return o.ImageRef, true
}

// HasImageRef returns a boolean if a field has been set.
func (o *Config) HasImageRef() bool {
	if o != nil && !IsNil(o.ImageRef) {
		return true
	}

	return false
}

// SetImageRef gets a reference to the given string and assigns it to the ImageRef field.
func (o *Config) SetImageRef(v string) {
	o.ImageRef = &v
}

// GetServiceAccountName returns the ServiceAccountName field value if set, zero value otherwise.
func (o *Config) GetServiceAccountName() string {
	if o == nil || IsNil(o.ServiceAccountName) {
		var ret string
		return ret
	}
	return *o.ServiceAccountName
}

// GetServiceAccountNameOk returns a tuple with the ServiceAccountName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Config) GetServiceAccountNameOk() (*string, bool) {
	if o == nil || IsNil(o.ServiceAccountName) {
		return nil, false
	}
	return o.ServiceAccountName, true
}

// HasServiceAccountName returns a boolean if a field has been set.
func (o *Config) HasServiceAccountName() bool {
	if o != nil && !IsNil(o.ServiceAccountName) {
		return true
	}

	return false
}

// SetServiceAccountName gets a reference to the given string and assigns it to the ServiceAccountName field.
func (o *Config) SetServiceAccountName(v string) {
	o.ServiceAccountName = &v
}

// GetResourceRequest returns the ResourceRequest field value if set, zero value otherwise.
func (o *Config) GetResourceRequest() PredictionJobResourceRequest {
	if o == nil || IsNil(o.ResourceRequest) {
		var ret PredictionJobResourceRequest
		return ret
	}
	return *o.ResourceRequest
}

// GetResourceRequestOk returns a tuple with the ResourceRequest field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Config) GetResourceRequestOk() (*PredictionJobResourceRequest, bool) {
	if o == nil || IsNil(o.ResourceRequest) {
		return nil, false
	}
	return o.ResourceRequest, true
}

// HasResourceRequest returns a boolean if a field has been set.
func (o *Config) HasResourceRequest() bool {
	if o != nil && !IsNil(o.ResourceRequest) {
		return true
	}

	return false
}

// SetResourceRequest gets a reference to the given PredictionJobResourceRequest and assigns it to the ResourceRequest field.
func (o *Config) SetResourceRequest(v PredictionJobResourceRequest) {
	o.ResourceRequest = &v
}

// GetImageBuilderResourceRequest returns the ImageBuilderResourceRequest field value if set, zero value otherwise.
func (o *Config) GetImageBuilderResourceRequest() ResourceRequest {
	if o == nil || IsNil(o.ImageBuilderResourceRequest) {
		var ret ResourceRequest
		return ret
	}
	return *o.ImageBuilderResourceRequest
}

// GetImageBuilderResourceRequestOk returns a tuple with the ImageBuilderResourceRequest field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Config) GetImageBuilderResourceRequestOk() (*ResourceRequest, bool) {
	if o == nil || IsNil(o.ImageBuilderResourceRequest) {
		return nil, false
	}
	return o.ImageBuilderResourceRequest, true
}

// HasImageBuilderResourceRequest returns a boolean if a field has been set.
func (o *Config) HasImageBuilderResourceRequest() bool {
	if o != nil && !IsNil(o.ImageBuilderResourceRequest) {
		return true
	}

	return false
}

// SetImageBuilderResourceRequest gets a reference to the given ResourceRequest and assigns it to the ImageBuilderResourceRequest field.
func (o *Config) SetImageBuilderResourceRequest(v ResourceRequest) {
	o.ImageBuilderResourceRequest = &v
}

// GetEnvVars returns the EnvVars field value if set, zero value otherwise.
func (o *Config) GetEnvVars() []EnvVar {
	if o == nil || IsNil(o.EnvVars) {
		var ret []EnvVar
		return ret
	}
	return o.EnvVars
}

// GetEnvVarsOk returns a tuple with the EnvVars field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Config) GetEnvVarsOk() ([]EnvVar, bool) {
	if o == nil || IsNil(o.EnvVars) {
		return nil, false
	}
	return o.EnvVars, true
}

// HasEnvVars returns a boolean if a field has been set.
func (o *Config) HasEnvVars() bool {
	if o != nil && !IsNil(o.EnvVars) {
		return true
	}

	return false
}

// SetEnvVars gets a reference to the given []EnvVar and assigns it to the EnvVars field.
func (o *Config) SetEnvVars(v []EnvVar) {
	o.EnvVars = v
}

// GetSecrets returns the Secrets field value if set, zero value otherwise.
func (o *Config) GetSecrets() []MountedMLPSecret {
	if o == nil || IsNil(o.Secrets) {
		var ret []MountedMLPSecret
		return ret
	}
	return o.Secrets
}

// GetSecretsOk returns a tuple with the Secrets field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Config) GetSecretsOk() ([]MountedMLPSecret, bool) {
	if o == nil || IsNil(o.Secrets) {
		return nil, false
	}
	return o.Secrets, true
}

// HasSecrets returns a boolean if a field has been set.
func (o *Config) HasSecrets() bool {
	if o != nil && !IsNil(o.Secrets) {
		return true
	}

	return false
}

// SetSecrets gets a reference to the given []MountedMLPSecret and assigns it to the Secrets field.
func (o *Config) SetSecrets(v []MountedMLPSecret) {
	o.Secrets = v
}

func (o Config) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Config) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.JobConfig) {
		toSerialize["job_config"] = o.JobConfig
	}
	if !IsNil(o.ImageRef) {
		toSerialize["image_ref"] = o.ImageRef
	}
	if !IsNil(o.ServiceAccountName) {
		toSerialize["service_account_name"] = o.ServiceAccountName
	}
	if !IsNil(o.ResourceRequest) {
		toSerialize["resource_request"] = o.ResourceRequest
	}
	if !IsNil(o.ImageBuilderResourceRequest) {
		toSerialize["image_builder_resource_request"] = o.ImageBuilderResourceRequest
	}
	if !IsNil(o.EnvVars) {
		toSerialize["env_vars"] = o.EnvVars
	}
	if !IsNil(o.Secrets) {
		toSerialize["secrets"] = o.Secrets
	}
	return toSerialize, nil
}

type NullableConfig struct {
	value *Config
	isSet bool
}

func (v NullableConfig) Get() *Config {
	return v.value
}

func (v *NullableConfig) Set(val *Config) {
	v.value = val
	v.isSet = true
}

func (v NullableConfig) IsSet() bool {
	return v.isSet
}

func (v *NullableConfig) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableConfig(val *Config) *NullableConfig {
	return &NullableConfig{value: val, isSet: true}
}

func (v NullableConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableConfig) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
