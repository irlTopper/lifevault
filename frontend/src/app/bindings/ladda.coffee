define ['knockout', 'ladda'], (ko, Ladda) ->

    ko.bindingHandlers.ladda =
        init: (element, valueAccessor) ->
            l = Ladda.create(element)
            ko.computed({
                read: () ->
                    state = ko.unwrap(valueAccessor())
                    if state
                        l.start()
                    else
                        l.stop()
                ,
                disposeWhenNodeIsRemoved: element
            })