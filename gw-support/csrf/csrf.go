// Package csrf contains functions for implementing CSRF protection through the use of a shared token secret between the
// client and the server. This way, we prevent allowing requests to the local server on certain endpoints via the
// browser, preventing cross site attacks.
// Inspired by houston-cli.
package csrf
