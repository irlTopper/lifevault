define [
    "knockout"
    "text!./template.html"
], (ko, homeTemplate) ->

    VM = ->
        # Vars
        app.mainHolderClass 'wrap-dashboard'

        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::OnShow = () ->
        $(".custom_scroll").mCustomScrollbar({theme: 'minimal-dark', scrollInertia: 0})
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
