define ['knockout', 'jquery'], (ko, $) ->

    ko.bindingHandlers.compareAt =
        init: (element, valueAccessor, allBindings) ->
            value = ko.unwrap(valueAccessor())
            previous = ko.unwrap(allBindings.get("previous"))
            buildComparison($(element), value, previous)

        update: (element, valueAccessor, allBindings) ->
            value = ko.unwrap(valueAccessor())
            previous = ko.unwrap(allBindings.get("previous"))
            buildComparison($(element), value, previous)

    buildComparison = (element, current, previous) ->
        element.removeClass("up down nochange")
        if current > previous
            element.addClass("up").html("+#{app.utility.CalculatePercentageChange(current, previous)}%")
        else if current < previous
            element.addClass("down").html("#{app.utility.CalculatePercentageChange(current, previous)}%")
        else
            element.addClass("nochange").html("0%")


    ko.bindingHandlers.difference =
        init: (element, valueAccessor, allBindings) ->
            value = ko.unwrap(valueAccessor())
            previous = ko.unwrap(allBindings.get("previous"))
            buildDifference($(element), value, previous)

        update: (element, valueAccessor, allBindings) ->
            value = ko.unwrap(valueAccessor())
            previous = ko.unwrap(allBindings.get("previous"))
            buildDifference($(element), value, previous)

    buildDifference = (element, current, previous) ->
        element.removeClass("up down nochange")
        if current > previous
            element.addClass("up").html("+#{current - previous}%")
        else if current < previous
            element.addClass("down").html("#{current - previous}%")
        else
            element.addClass("nochange").html("0%")

