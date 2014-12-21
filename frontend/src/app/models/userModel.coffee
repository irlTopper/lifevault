###
    User model
###
define ['knockout', 'knockout-mapping'], (ko, mapping) ->
    User = (data) ->
        # Map the data across
        mapping.fromJS data, {}, this

        ###
        Extras and defaults
        ###
        # The googlePlusURL field may not have been return - set it to empty string
        if typeof @googlePlusURL is 'undefined' then @googlePlusURL = ''

        return

    # Deliberately not a computed - no need
    User::FullName = () ->
        return @firstName() + ' ' + @lastName()

    User.mapping =
        'user':
            'create': (options) ->
                return new User(options.data)
        'users':
            'key': (data) ->
                ko.utils.unwrapObservable(data.id)
            'create': (options) ->
                return new User(options.data)

    return User