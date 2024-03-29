package manager

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

//this package (substitution) is meant to facilitate variable substitution in command lines

func installCommandVariableSubstitution(cmdLine, pkg string, sudo bool) string {
	cmdLine = strings.Replace(cmdLine, "${pkg}", pkg, -1)
	return replaceSudo(cmdLine, sudo)
}

func replaceSudo(cmdLine string, sudo bool) string {
	if sudo {
		cmdLine = strings.Replace(cmdLine, "${sudo}", "sudo", -1)
		cmdLine = strings.Replace(cmdLine, "${SUDO}", "sudo", -1)
	} else {
		cmdLine = strings.Replace(cmdLine, "${sudo}", "", -1)
		cmdLine = strings.Replace(cmdLine, "${SUDO}", "", -1)
	}
	return strings.TrimSpace(cmdLine)
}

// injectVars first tries to replace all ${SH} style variables with the envy configuration values,
// then with any environment variables.
func injectVars(cmdLine string, vars envVariables, sudo bool) string {
	//need to do sudo expansion first, since it's a special case
	cmdLine = replaceSudo(cmdLine, sudo)

	for k, v := range vars {
		cmdLine = strings.Replace(cmdLine, fmt.Sprintf("${%v}", strings.ToUpper(k)), v, -1)
		cmdLine = strings.Replace(cmdLine, fmt.Sprintf("${%v}", strings.ToLower(k)), v, -1)
	}

	//now search for any leftover requests intended to get environment variables
	//regular expressions...ewwwww you say.... But I like them!
	reg := regexp.MustCompile(`\${(\w+)}`)
	matches := reg.FindAllStringSubmatch(cmdLine, -1)
	if matches == nil {
		return cmdLine
	}
	for _, match := range matches {
		//try to get the environment variable defined here
		v := os.Getenv(match[1])
		if v == "" {
			//TODO (@morgan): possibly warning here that a leftover variable expression did not get expanded
			continue
		}
		cmdLine = strings.Replace(cmdLine, match[0], v, -1)
	}
	return cmdLine
}

func clean(input string) string {
	input = strings.TrimSpace(input)
	return strings.ToLower(input)
}

// if CLI arguments are supplied, they over-ride package/installer preferences
func determineSudo(config RunConfig, installer *Installer) bool {
	if strings.ToLower(config.Sudo) == "true" || strings.ToLower(config.Sudo) == "t" {
		return true
	}
	if strings.ToLower(config.Sudo) == "false" || strings.ToLower(config.Sudo) == "f" {
		return false
	}
	if installer == nil {
		return false
	}
	return installer.Sudo
}
