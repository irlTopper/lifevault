define [
    "knockout"
    "text!./template.html"
    "knockout-mapping"
    "helpdocModel"
    "customerModel"
    "ticketListItemModel"
    "articleListItemModel"
    "lodash"
], (ko, templateMarkup, mapping, helpdocModel, customerModel, ticketListItemModel, articleListItemModel, _) ->
    VM = (params) ->
        # Vars
        @readyToShow = ko.observable(false)
        @searchState = ko.observable("start")


        app.InitOnShow(@)# Sets-up the OnShow() function
        @readyToShow(true)
        return


    VM::OnShow = () ->
        # Focus on the search input by default
        @focusSearch()
        return

    VM::onSubmitSearch = () ->
        @searchTerm( @newSearchTerm() )

        app.GoTo "#search/" + @selectedTab().URLpart + "/" + @searchTerm()

        @doSearchAgain()


    # This runs when the component is torn down. Put here any logic necessary to clean up,
    # for example cancelling setTimeouts or disposing Knockout subscriptions/computeds.
    VM::dispose = () ->
        if subsc_ticketLastUpdatedSelected? then subsc_ticketLastUpdatedSelected.dispose()
        if @subsc_customerlastUpdatedSelected? then @subsc_customerlastUpdatedSelected.dispose()
        if @subsc_helpdocsLastUpdatedSelected? then @subsc_helpdocsLastUpdatedSelected.dispose()
        if @subsc_helpdocsSelectedSites? then @subsc_helpdocsSelectedSites.dispose()
        if @subsc_helpdocsSelectedUsers? then @subsc_helpdocsSelectedUsers.dispose()
        if @subscr_selectedTag? then @subscr_selectedTag.dispose()
        return


    return {
        viewModel: VM
        template: templateMarkup
    }