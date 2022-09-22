/*
Package configo is a configuration manager. 
Populate your configuration struct from flags, env vars or external providers.
For example read from a secret manager and refresh the value every 12H. Or override with local value when running locally.

Features:
 
  - Read from flags, env files, local files, remote config
  - Easy use of custom types
  - Declarative style OR/AND struct tags style
  - SubCommands with persistent or local flags.
  - Parameter options:
    - Mandatory values
    - Enum values (list of allowed values, implement interface.Values() or interface.List())
    - Custom flag name or envvar name.
    - Value validation
    - Description
    - Examples
    - Default value
    - Exclusive params (either param1 or param2 but not both)
  
  - No external libraries
  
  - Refresh conf
  - Low footprint once the init is done


Limitations:

 - Every input is always a string and must be transformed. Empty strings are skipped.
 - The reading priority is always the same.

Priorities:

The value will be set in this order, each step overriding the previous:

 1) Default in the code
 2) Loader (user defined function, read for local file, secret manager...)
 3) Env Var
 4) Command line flags

For example a param has a synchronization configuration AND an env var value. 
Then the loader won't be used at all, the env var value will be kept. 

Helpers:

This code also provides some helpers:
 - SecretRotation to help with rotating secrets, for example a consumer calling a service requiring an API secret.
 - AwsSecretManager to fetch the configuration in https://aws.amazon.com/secrets-manager/ (This module has additional dependencies)

Integration:

Real life example of use case. Input configuration for a server web. It requires some general configuration and also some API endpoints are protected by a secret. The configuration (and the secrets) are all in the AWS secret manager, or are overridden locally on this instance using flags or env var.

 1) Define a first minimal configuration for AWS secrets.
 => this can be provided by many ways. Recommended way would be by instance role or by env var. (then no configuration needed at all in the code).

 2) Define the full configuration
 => Some parameters are contants once loaded (web server timeout, table name ...)
 => Some parameters can be sync regularly (log level, ...)
 => Some secrets are sync regularly and use the secretrotation package:

Because there is a synchronization, use the lock when reading values to avoid race condition.

*/
package configo
