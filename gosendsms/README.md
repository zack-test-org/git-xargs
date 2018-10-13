# gosendsms

Simple command line utility for sending a text message via
twilio.


## Prerequisite

You must have a Twilio account and phone number with SMS capabilities. Follow [their free trial guide](https://www.twilio.com/docs/usage/tutorials/how-to-use-your-free-trial-account) for instructions on setting one up.

Note that if you decide to upgrade, it will cost you $1 to
maintain the phone number and 0.7 cents ($0.007) per message
sent.

Then, you need to set the following environment variables:

```
TWILIO_ACCOUNT_ID
TWILIO_AUTH_TOKEN
TWILIO_SEND_FROM_NUMBER
TWILIO_SEND_TO_NUMBER
```


* `TWILIO_ACCOUNT_ID` and `TWILIO_AUTH_TOKEN` refer to your
  Twilio account SID and auth token, which are used for
  authenticating requests to Twilio. See [the relevant support
  topic](https://support.twilio.com/hc/en-us/articles/223136027-Auth-Tokens-and-how-to-change-them)
  for instructions on how to find them.
* `TWILIO_SEND_FROM_NUMBER` refers to the twilio number you registered above.
* `TWILIO_SEND_TO_NUMBER` refers to the twilio number you want to send the message to.

It is best to store these in `pass`, using `pass insert -m
twilio`. Then, you can pull these out using `eval $(pass
twilio)`.


## Usage

- `go build gosendsms.go`
- `./gosendsms send message` => You should receive a text message to the
  phonenumber set under `TWILIO_SEND_TO_NUMBER` from the phone number set under
  `TWILIO_SEND_FROM_NUMBER`.
- If you add this to your `PATH`, you can now use this as a notification for
  when a long running task finishes. For example, `terraform apply; gosendsms
  deployment completed`. You can even include the exit code in the message to
  know if the task succeeded. `terraform apply; gosendsms deployment completed $?`.
