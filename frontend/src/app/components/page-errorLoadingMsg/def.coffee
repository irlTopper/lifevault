define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->
    VM = (params) ->

        # Setup so if users navigates away, this message tears down
        @routeSub = app.currentRoute.subscribe () =>
            app.pageLoadingErrorXHR(null)

        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    # This runs when the component is torn down. Put here any logic necessary to clean up,
    # for example cancelling setTimeouts or disposing Knockout subscriptions/computeds.
    VM::dispose = () ->
        @routeSub.dispose()
        return

    VM::TryAgain = ->
        app.pageLoadingErrorXHR(null)
        return



    return {
        viewModel: VM
        template: templateMarkup
    }