# Authentication #

## Device ID ##
An encrypted browser cookie is used to store an random device-id.
This is assigned on first used, or after browser history was cleared.

This is used to identify a session, but is not a uniq session ID, because multiple devices for the same used can be associated with the same session (see later).

If the deviceID is not associated with a session, the login process starts.

## Login Process ##
1. Store the requested page in the cookie (session not yet available)
1. Display login-email-form to ask for email
1. Send OTP to the email address
1. Display login-otp-form to ask for the OTP
1. Validate OTP
1. Proceed to requested page (retrieved from the cookie)

## Session ##
After login with a new email address not previously used, a new session is created and the device-id is added to the session as a lookup key.
If a session already exists for the email address, the device-id is added to it and both devices use the same session.

There can be only one session per email - shared across all devices where the email logs in.
Multiple device ids can be associated with the same session.

The device activity is tracked in the session and the device ID may be expired if necessary which will require a new login from that device.
The device ID may also be removed manually from another device if a device was lost/broken/stolen/...

The device ID is an UUID encrypted into the cookie. If the cookie is copied across to another device, it too will have access to the session.
If device hardware can be identified, one can prevent this, but not seen as a risk at the moment as the user needs to be careless or coorporative for this to be possible and this system does not require the strictest access control like a banking app.
