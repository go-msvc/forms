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

- 2023-03-11
- Created micro-service to manage forms (using simple local file system)
- Added micro-service call to web to fetch form from the micro-service
- Created micro-service operations to store documents
- Displaying other types of fields
- Submit documents into micro-service
- Added support for markdown in titles and descriptions
- Added campaign to the service and to web, but web wait for redisClient to send a notification

## Next ##
- Add campaigns to control document substitution and notification sending ...
- Consume notifications
- Template: Field description as hint when hover over field
- Form Template must use java script to validate forms while filled in
- Form validate: all fields must have unique names in scope
- Allow existing document to be opened for editing and submit new version of document
- Add basic field validation into the form and template (e.g. date min/max already added)
- Add basic display options
- Support for user scripts - compile into go code? or see if can execute python code? Handle exceptions!

## Apps ##
- campaigns: form start/end-time, limited submits, ... build this outside the form, e.g. in a "campaign" and expose the form then using a name, rather than the id, and let the URL be like a multi-folder path name to make hierarchies

- Move data into a database - SQL for foreign key and lookup and object for the nested data


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
