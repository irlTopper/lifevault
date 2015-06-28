define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->
    VM = (params) ->
        app.mainHolderClass 'wrap-profile'

        # Vars
        @userId = app.loggedInUser.id()

        # Find the user that matches the one on the URL, or redirect to dashboard if not found
        @user = app.loggedInUser
        if not @user?
            # No match found - users are essential data - let MissingEssentialData() handle it
            app.MissingEssentialData()
            return

        #Define the side navigation
        @sideNav = [
            {
                name: "Profile"
                routePart: "profile"
                icon: "icon-mine"
            }
        ]


        @sideNav.push({
            name: "Reminders"
            routePart: "reminders"
            icon: "icon-link"
        })

        @sideNav.baseURL = "#myprofile/" #should be #users/333
        @sideNav.routePartName = "view" #This should match the name the router uses for this part
        ###
        {
            name: "Quality Control"
            routePart: "autobcc"
            icon: "icon-bbc"
        }
        ###

        # Ensure valid route
        validOpts = []
        validOpts.push nav.routePart for nav in @sideNav
        app.ensureValidRoute('view',{vm:this,validOpts:validOpts,rememberIn:"lastUserView",onlyIfSame:"pageTab"})


        # Setup the selectedPanel observable
        @selectedPanel = app.ko.computed =>
            return 'section-user-' + app.currentRoute().view

        app.InitOnShow(@)# Sets-up the OnShow() function
        return


    # This runs when the component is torn down. Put here any logic necessary to clean up,
    # for example cancelling setTimeouts or disposing Knockout subscriptions/computeds.
    VM::dispose = () ->
        app.removeValidRouteSubscriptions(this)
        return

    return {
        viewModel: VM
        template: templateMarkup
    }
