define [
    "knockout"
    "text!./template.html"
], (ko, homeTemplate) ->

    VM = ->
        @isLoaded = app.ko.observable false


        app.mainHolderClass 'wrap-dashboard'

        # Setup the nav
        @subnav = [
            { name: "Inboxes", routePart: "inboxes" }
            { name: "Metrics", routePart: "metrics" }
        ]
        @subnav.baseURL = "#" + app.currentRoute().page + "/" #should be #dashboard
        @subnav.routePartName = "dashboardPageTab" #This should match the name the router uses for this part


        # Ensure valid route
        validOpts = []
        validOpts.push nav.routePart for nav in @subnav
        app.ensureValidRoute('dashboardPageTab',{vm:this,validOpts:validOpts,rememberIn:"lastDashboardPageTab",onlyIfSame:"page"})


        @isLoaded true
        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::OnShow = () ->

        return

    # This runs when the component is torn down. Put here any logic necessary to clean up,
    # for example cancelling setTimeouts or disposing Knockout subscriptions/computeds.
    VM::dispose = () ->
        app.removeValidRouteSubscriptions(this)
        return


    return {
        viewModel: VM
        template: homeTemplate
    }
