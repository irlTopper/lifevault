define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->
    VM = (params) ->
        @nav = params.nav

        i = 0
        while i < @nav.length
            #Build the full urls for each sub-nav item
            @nav[i].url = @nav.baseURL + @nav[i].routePart
            i++
        return

        # Ensure valid route
        validOpts = []
        validOpts.push nav.routePart for nav in @nav
        app.ensureValidRoute('view',
            vm: this
            validOpts: validOpts
            rememberIn: "lastUserView"
        )
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
