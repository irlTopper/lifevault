# Note that this file is also read by the build script
components = [
    'page-loading'

    # Pages
    'page-journal'
    'page-calendar'
    'page-errorLoadingMsg'
    'page-login'
    'page-reset-password'
    'page-forgot-password'
    'loadingIndicator'
    'page-search'
    'page-myprofile'
    'section-blankslate'
    'section-pagination'
    'section-user-profile'
    'modal-confirm'
    'modal-imageupload'
    'modal-keyboard-shortcuts'
    'modal-prompt'
    'pageLHS-settings-nav'

    'titlebar'
    'widget-uiMessage'

    'nav-main'
]


if window?
    define ['knockout'], (ko) ->
        for component in components
            ko.components.register component, { require: "app/components/#{component}/def" }