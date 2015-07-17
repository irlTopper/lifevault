define [
    "knockout"
    "text!./template.html"
], (ko, homeTemplate) ->

    VM = ->
        # Vars

        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::OnShow = () ->

        return

    # This runs when the component is torn down. Put here any logic necessary to clean up,
    # for example cancelling setTimeouts or disposing Knockout subscriptions/computeds.
    VM::dispose = () ->

        return


    return {
        viewModel: VM
        template: homeTemplate
    }
