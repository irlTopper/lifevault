define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->

    VM = (params) ->
        @title = params.title
        @subnav = params.subnav
        @routePartName = params.routePartName

        if @subnav?
            for nav in @subnav
                #Build the full urls for each sub-nav item
                nav.url = @subnav.baseURL + "/" + nav.routePart

        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::OnClickGettingStarted = () ->
        if app.loggedInUser.onboarding.dismissedWelcome()
            app.UpdateOnboarding 'dismissedWelcome', false

    return {
        viewModel: VM
        template: templateMarkup
    }
