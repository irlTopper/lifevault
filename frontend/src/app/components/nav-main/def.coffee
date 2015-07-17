define [
    "knockout"
    "text!./template.html"
], (ko, template) ->
    VM = (params) ->
        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    return {
        viewModel: VM
        template: template
    }
