TeamworkDesk Setup Guide
================

- Install Revel by running `go get github.com/revel/revel`
- Install Revel CMD by running `go get github.com/revel/cmd/revel`
- Make sure your Go bin path is in your PATH environment variable (located in GOPATH/bin)
- Clone this repository by running `go get github.com/Teamwork/TeamworkDesk`
- Create the config file: conf/app.conf
	- Copy/rename sample.app.conf to app.conf
	- See http://revel.github.io/manual/appconf.html
- Setup database:
	- Create databases `teamworkdesk_shard6` and `teamworkdesk_temp` on your local machine
	- Import their respective SQL dumps ([found here](http://digitalcrew.teamwork.com/projects/38791/files?catId=44988))
	- Add them as datasources in `teamworkpm_master.datasources`
	- Add the user you use to login to local TWPM to `teamworkdesk_shard6.users`
	- Add `companyLoginBackgroundCOlor VARCHAR(6) NULL DEFAULT ''` column to `teamworkpm_shard6.companies`
	- Add `12 hours` and `24 hours` names to `timeFormatStringDisplayNameEN` column in `teamworkpm_master.timeformatstrings`
	- Add `inactive` option to `emailForwardingState` enum in `teamworkdesk_shard6.inboxes`
- Setup Redis as a service
	- Go to `Dropbox/TeamworkPMSharedFiles/Setup/redis`
	- Run `redis-server --service-install redis.windows.conf --loglevel verbose`
	- Open up `Services` and start Redis.

# The Robot
Each TeamworkDesk database (eg. teamworkfdesk_shard6) has to have a robot user.
This is a reference in the database table to user 999999999, "Mr Robot".  We have set this id
as a constant here- https://github.com/irlTopper/ohlife2/blob/master/app/controllers/appController.go#L17

In future development adding this user will be automated, but for now, please check all teamworkfdesk_shard(n) databases and ensure that you have a user record with the ID of 999999999.

This fictional user is used for referential integrity when there are anonymous
jobs such as new tickets being created from Mailgun and we have who-did-it fields
like "createdByUserId".
(Note we could have gone with a null-able column for who-did-it tracking columns
but that would be slower and messier to work)

# Email Notifications
(Lack of better place to put this right now, but I would like to develop a bit of a comprehensive list of "how things in the app work and why")

Email notifications are sent to agents (desk users) on the following actions-
* OnNewTicket
* OnTicketAssigned
* OnTicketAssignedToSomebodyElse
* OnCustomerRepliesToNewTicket
* OnCustomerRepliesToMyTickets
* OnCustomerRepliesToSomebodyElsesTicket
* OnOtherUserRepliesToNewTicket
* OnOtherUserRepliesToMyTicket
* OnOtherUserRepliesToSomebodyElse

These settings can be enabled per inbox, however, if there exists an entry for inboxID = 0, then that means we will use the same settings for all inboxes.

# Coding Standards
### Struct Mappings
Gorp can be a bit of a pain sometimes.  We have found that we need to have different structs to represent slight variations of the data (most commonly json and db differences).  In order to keep things well organized and quick to decipher, we have created a standard of `NameVersion`.  Two examples of this standard are `UserDb` and `UserJSON`.



# Dev Tips
- **Epic** Install [SublHandler](https://github.com/ktunkiewicz/subl-handler) (Can jump straight to file by clicking an error)
- Sublime Text 2
-- Install Gosublime
-- Install the goimports tool: http://blog.campoy.cat/2013/12/integrating-goimports-with-gosublime-on.html
- Chrome extenstion "Easy auto refresh"
-- During dev it is often useful to use this - set it to update every 3 to 5 seconds while
-- you break/fix stuff. You see from the console when everything is working or where errors are.
- Revel and Go are as scary as they seem
- Go makes copies of variables by default - this can be a big gotcha when starting. Use &var.



# A Tale of two codebases
We have both the frontend and backend here in one codebase.

The frontend exists under [/frontend](https://github.com/irlTopper/ohlife2/tree/master/frontend) and comes with it's own [readme](https://github.com/irlTopper/ohlife2/blob/master/frontend/README.md)
