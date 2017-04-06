package utils

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/eagerio/Stout/src/providers"
	"github.com/eagerio/Stout/src/providers/providermgmt"
	"github.com/imdario/mergo"
	yaml "gopkg.in/yaml.v1"
)

func ErrorMerge(str string, err error) error {
	return errors.New(str + " " + err.Error())
}

func PanicsToErrors(debugMode bool, f func() error) (err error) {
	if !debugMode {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				err, ok = r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
			}
		}()
	}

	return f()
}

func LoadEnvConfig(filledProfile providers.EnvHolder) (profile providers.EnvHolder, err error) {
	t := providers.ConfigHolder{}

	configProvided := true
	if filledProfile.GlobalFlags.Config == "" {
		configProvided = false
		filledProfile.GlobalFlags.Config = "./config.yaml"
	}

	data, err := ioutil.ReadFile(filledProfile.GlobalFlags.Config)
	if err != nil {
		if configProvided {
			return filledProfile, err
		}

		return filledProfile, nil
	}

	err = yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		return
	}

	if filledProfile.GlobalFlags.Debug {
		fmt.Println("Original config file used:")
		d, err := yaml.Marshal(&t)
		if err != nil {
			return filledProfile, err
		}
		fmt.Println(string(d))
	}

	envProvided := true
	if filledProfile.GlobalFlags.Env == "" {
		envProvided = false
		filledProfile.GlobalFlags.Env = "default"
	}

	profile, ok := t[filledProfile.GlobalFlags.Env]
	if !ok {
		if envProvided {
			return filledProfile, errors.New("Env provided does not exist")
		}

		fmt.Println("Env not provided, default env not found, ignoring config file.")
		return filledProfile, nil
	}

	if filledProfile.GlobalFlags.Debug {
		fmt.Println("Config file env profile used:")
		d, err := yaml.Marshal(&profile)
		if err != nil {
			return filledProfile, err
		}
		fmt.Println(string(d))
	}

	for provider, _ := range profile.ProviderFlags {
		originalProviderObj := providermgmt.ProviderList[provider]
		providerObj := providermgmt.ProviderList[provider]

		providerStr, err := yaml.Marshal(profile.ProviderFlags[provider])
		if err != nil {
			return filledProfile, err
		}

		err = yaml.Unmarshal(providerStr, providerObj)
		if err != nil {
			return filledProfile, err
		}

		err = mergo.MergeWithOverwrite(originalProviderObj, providerObj)
		if err != nil {
			return filledProfile, err
		}

		providermgmt.ProviderList[provider] = originalProviderObj
	}

	if filledProfile.GlobalFlags.Debug {
		fmt.Println("Config file profile before merge:")
		d, err := yaml.Marshal(&profile)
		if err != nil {
			return filledProfile, err
		}
		fmt.Println(string(d))

		fmt.Println("Command line flag profile (including defaults) before merge:")
		d, err = yaml.Marshal(&filledProfile)
		if err != nil {
			return filledProfile, err
		}
		fmt.Println(string(d))
	}

	err = mergo.MergeWithOverwrite(profile.GlobalFlags, filledProfile.GlobalFlags)
	if err != nil {
		return filledProfile, err
	}
	err = mergo.MergeWithOverwrite(profile.CreateFlags, filledProfile.CreateFlags)
	if err != nil {
		return filledProfile, err
	}
	err = mergo.MergeWithOverwrite(profile.DeployFlags, filledProfile.DeployFlags)
	if err != nil {
		return filledProfile, err
	}
	err = mergo.MergeWithOverwrite(profile.RollbackFlags, filledProfile.RollbackFlags)
	if err != nil {
		return filledProfile, err
	}

	if filledProfile.GlobalFlags.Debug {
		fmt.Println("Final config file used:")
		d, err := yaml.Marshal(&filledProfile)
		if err != nil {
			return filledProfile, err
		}
		fmt.Println(string(d))

		fmt.Println("Provider flags used:")
		d, err = yaml.Marshal(&providermgmt.ProviderList)
		if err != nil {
			return filledProfile, err
		}
		fmt.Println(string(d))
	}

	return profile, nil
}
