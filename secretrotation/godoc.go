/*
Package secretrotation helps with secret rotation when a service must accept old and new secrets for a time.

  - SecretHolder is a DB service that provides the secret. For example a AWS Secret Manager service. This should provide 3 secrets (default: comma separated).
  - Provider is for example a web server validating secret in incoming API calls. At any time, it always accepts any of the 3 secrets.
  - Consumer is for example an application calling the web server. It is always sending the secret in the middle.

The reason for the 3 secrets is we always want a valid secret, given:
 1. In the SecretHolder, the secret rotation can happen anytime (but not too frequent).
 2. The providers can load the secrets anytime.
 3. The consumers can load the secrets anytime.

Assertions:
  - The refresh rate on the Providers and Consumers is faster than the secret rotation frequency in the secret holder.
  - We have 3 secrets: {PREVIOUS,CURRENT,PENDING}. This is AWS design but that seems robust.

See: https://docs.aws.amazon.com/secretsmanager/latest/userguide/getting-started.html

Alternative similar design:
  - store only 2 secrets in the SecretHolder but with the change time, and share rotation duration details with the consumer. (A lot more complicated)
*/
package secretrotation
