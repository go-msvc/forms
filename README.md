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
- Moved form.tmpl hidden fields into internal context/cookie...
- Using internal session stored in-memory in the micro-service
- Added basic field validation into the form description and template (e.g. date min/max already added) but not yet applied

- 2023-03-16
- Auth is basically working with internal session data and device ID, but error handling is poor and no links to go back to login/out... need to add nav bar to top of page to show email when logged in and link to logout etc...

- 2023-03-17
- Added topnav to control login and logout - working
- Nav bar is currently static - need to respond to page

## Next ##
- Test well for normal users who starts with campaign link
- Show recent campaigns in user's home
- Show my campaigns
- Control nav bar items

- User home owning a campaign: Show entries, open them and select action, and list of entries in different states.
- Other user also to see entries when in processing group
- Option to add other user or remove users
- Show list of actions/queues and entries in queues

- Later: Give user opportunity to edit profile with list of persons in own profile
- Then option in a form to select a person from your list...
    In the entry - distinguish between submitted and person being entered
    Allow forms with multiple persons
    Do validation on selected person's fields...

- Identify the user and store a profile - even if empty
    but indicate if email is not set so that user can recover profile when moving to other device
    associate multiple devices to a profile

- Let submit store client token, so that messages can be queued for a client to review without having to login
    client can also open from link sent in email...

- Busy with consumer...
- Consume notifications to do:
    - add docs to internal lists in a db, e.g. forReview, accepted, rejected, returned, ...
    - display list of documents in a list (may be add to list without need for a service)
    - let user see list of docs, open each, then do something
    - ...

- Show queue of documents to a user/group... for review
- Show actions that user can take
- Create a new document from an old document, with some fields copied and some edited... e.g. let reviewer add comments and mark who did the review when... then queue for another group/user


- Block campaign when REDIS KEY is longer than N - i.e. no longer processed
- Template: Field description as hint when hover over field
- Form Template must use java script to validate forms while filled in
- Allow existing document to be opened for editing and submit new version of document
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
