define ['knockout'], (ko) ->

    # Return a minutes value as an hours and minutes string eg "4m" or "3h 45m"
    ko.bindingHandlers.hoursAndMins =
        update: (element, valueAccessor, allBindingsAccessor) ->
            ko.bindingHandlers.html.update element, ->
                value = (ko.utils.unwrapObservable(valueAccessor()) or 0)
                m = value % 60
                h = ( value - m ) / 60
                r = 0
                if m > 0 then r = r + Math.floor(m) + "m"
                if h > 0
                    if m > 0 then r = " " + r
                    r = "" + h + "h" + r
                return r

    ko.bindingHandlers.fadeVisible =

        update: (element, valueAccessor) ->
            # Whenever the value subsequently changes, slowly fade the element in or out
            value = valueAccessor()
            if ko.unwrap(value) then $(element).fadeIn() else $(element).fadeOut()
            return
        init: (element, valueAccessor) ->
            # Initially set the element to be instantly visible/hidden depending on the value
            value = valueAccessor()
            $(element).toggle ko.unwrap(value)
            # Use "unwrapObservable" so we can handle values that may or may not be observable
            return

    return