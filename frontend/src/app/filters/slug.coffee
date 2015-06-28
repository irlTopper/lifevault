define ['knockout', 'punches'], (ko) ->

    ko.filters.slugify = (val) ->
        val = ko.unwrap(val)
        val.toString()
            .toLowerCase()
            .replace(/\s+/g, '-')           # Replace spaces with -
            .replace(/[^\w\-]+/g, '')       # Remove all non-word chars
            .replace(/\-\-+/g, '-')         # Replace multiple - with single -
            .replace(/^-+/, '')             # Trim - from start of text