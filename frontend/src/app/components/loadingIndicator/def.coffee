define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->
    VM = (params) ->
        @readyToShow = ko.observable(true)
        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::OnShow = () ->

    return {
        viewModel: VM
        template: templateMarkup
    }