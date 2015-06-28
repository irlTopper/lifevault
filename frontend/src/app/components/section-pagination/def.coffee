define [
    "knockout"
    "text!./template.html"
], (ko, templateMarkup) ->
    VM = (params) ->

        @single = params.single
        @plural = params.plural
        @count = params.count
        @items = params.items
        @pageSize = params.pageSize
        @isLoading = params.isLoading
        @onClickNextPage = params.onClickNextPage

        return

    return {
        viewModel: VM
        template: templateMarkup
    }
