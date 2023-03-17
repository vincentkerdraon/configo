/*
Package awssecretmanager helps loading a secret from https://aws.amazon.com/secrets-manager/

Helper for the default format available from the console:
  - plain text
  - JSON.

Rotation state:
  - disable: there is only one value.
  - enable: a lambda is rotating the secret. Retriving values for the stages: Previous + Current + Pending

When the rotation is disabled, this package will return the Current value for all the stages.

Check also the go lambda package to rotate the secret.
*/
package awssecretmanager
