define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->
    VM = (params) ->
        @title = params.title
        @text = params.text
        @icon = params.icon
        @onClick = params.onClick
        @clickText = params.clickText
        app.InitOnShow(@)# Sets-up the OnShow() function
        return

    return {
        viewModel: VM
        template: templateMarkup
    }
