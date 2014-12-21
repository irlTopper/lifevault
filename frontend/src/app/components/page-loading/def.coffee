define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->
    VM = (params) ->
        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    # This runs when the component is torn down. Put here any logic necessary to clean up,
    # for example cancelling setTimeouts or disposing Knockout subscriptions/computeds.
    VM::dispose = ->

    return {
        viewModel: VM
        template: templateMarkup
    }