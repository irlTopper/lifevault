define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->
    VM = (params) ->
        @isLoaded = ko.observable(true)
        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    VM::OnShow = () ->

    return {
        viewModel: VM
        template: templateMarkup
    }