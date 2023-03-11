# Forms #

Part of go-msvc. This includes:
- Web API to display forms
- Micro-service to provide/manage the forms
- Micro-service to manage submitted form data

Note: to access the forms outside, need generic go-msvc api gateway to get it from the micro-services

## Status ##

- 2023-03-09
- Added basic web app with a login page template copied from https://www.w3schools.com/
- Not yet processing the login info, but ok for now...
- Created a form definition in forms.Form{}
- Defined a testForm["1"] which can be viewed with /form/1
- Form can also be submitted, but data not processed/stored, just debug logged.
- Currently only displaying simple text inputs - regardless of type
- All sections are rendered on one page


## Next ##
- Form submit mechanism - need to store the submitted data somewhere as a JSON doc
- then validation and a few more types of fields (int at least and choices/selections)
- Link that to the micro-service to get the form spec and to store the form data
- Form cancel must go somewhere or just remove the button? But rather keep it, then also add buttons to [Save] or [Undo] changes or [Reset], or [Start another entry] etc...

## Later ##
- Allow user to edit a submitted doc (option on form and doc to allow/block editing)
- Would be nice if comments can be added to each doc field when sent for review

- Better display in the form sections, nav between sections, ...
- Implement Form lists and tables and sub sections

- Would be nice to save work-in-progress without submitting and remind and cleanup if forgotten or too late to submit etc...

- Send submit/edit/delete/... notification from the back-end so that another micro-service can process it.
    (e.g. REDIS so queue persist if processor is down)

- See if can store user data on the device to make form entry easy
- But first try to do all submits/edits without user identification
