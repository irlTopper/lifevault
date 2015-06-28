# Note that this file is also read by the build script
components = [
    'page-loading'

    # Pages
    'page-dashboard'
    'page-errorLoadingMsg'
    'page-login'
    'page-reset-password'
    'page-forgot-password'
    'loadingIndicator'
    'page-plan'
    'page-search'
    'page-myprofile'
    'page-journal'
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