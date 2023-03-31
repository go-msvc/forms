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

- 2023-03-19
- Written focus section below... need to keep focus on what is needed to work best
- Need to focus on back-end first!
- Added user home and campaing page - but really basic and user home does not show real data yet.
    because not working on front-end for now.

## Next (UI) ##
- 2023-03-30:
    - Show Campaign: Chrome Incognito Window:
            http://localhost:8080/user/campaign/8df7997b-4a90-4784-a496-c0ec453c10f2#about
    - With incognito:
        - Ask for email -> then OTP -> then show form
            and included confirmed email in submission
    - Include confirmed email in each submission
    - For return user:
        - Not yet submitted:
            - show form immediately
        - Already submitted:
            - show list of existing submissions and option to submit another

    - Remove cookies/session data after login, e.g. otp values can be cleared
    - Improve form validation in the front-end...
    - Later for return user - show returned forms + comments to edit and submit updates

## Then (BE) ##
- ALWAYS - BACK-END rather than front-end
    - Front-end can be improved later
    - For now, all features must be available in the back-end first.

- Service DB Update
    - store relational data for forms/campaigns stored as JSON on disk
    - need a sql table of key data for lookups to find list of ids to manage
    - storage currently on disk can later move to an object db, no need to move yet.
        - 2023-03-30: fixed go-msvc/redis-utils now working keyvalue store to use

- Basic submissions:
    - show your campaigns in user's home
    - backend create a campaign with no automation owned by a user
    - share a link so others can fill and submit
    - show submitted forms
    - download data as CSV/JSON
- Automation:
    - put submissions in processing queue
    - show queue
    - select other queue and move the doc
- Reject/Review
    - Send form back to user for editing (email) with a comment - track the progress steps
    - Allow to edit until re-submitted then back in queue
    - Show review notes to the user... later can show per field or only show fields to be reviewed
- Scripted automation
    - Scripted validation of field values - including api calls and logic

- Before publish
    - Throttle queries from same source for DOS attacks
    - Limit free-tier nr of forms/campaigns/submissions per users
    - Allow to change the limits per user

- Protected Details
    - Allow users to enter profile details and enter a form as a person without sharing your details
    - So the submission only contains your ID & Name, which the campaign driver can use to contact you
        by sending messages to you which you can decide to ignore or read.
    - Then show messages to user after login
    - Later can push to mobile
    - Allow user to block filter senders. Auto allow user to send to you after you submitted.

- Multi Profile
    - Allow creation of profiles for other people (like your children)
    - Submit forms on their behalf - again without charing the details.
    - Let campaign ask for personal details they need
    - Let submitted approve sending of selected info
    - Show reply messages per profile in your inbox


## Bugs ##

- User's home entered by URL (after login) does not show user details in NAV bar... using different handler - need to get the data in this case - prevents user from logout!




## Next ##

- Test well for normal users who starts with campaign link
- Show recent campaigns in user's home - for now created in back-end.
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
