/*
 * This file is highly based on the example for theme handling provided by Bootstrap, with local modifications:
 * https://getbootstrap.com/docs/5.3/customize/color-modes/#javascript
 * Copyright 2011-2023 The Bootstrap Authors
 * Licensed under the Creative Commons Attribution 3.0 Unported License.
 */

export enum Theme {
    Light = 'light',
    Dark = 'dark',
    Auto = 'auto',
}

function getStoredTheme(): Theme | null {
    return localStorage.getItem('theme') as Theme
}

function setStoredTheme(theme: Theme) {
    localStorage.setItem('theme', theme)
}

export function getPreferredTheme(): Theme {
    const storedTheme = getStoredTheme()
    if (storedTheme) {
        return storedTheme
    }

    return window.matchMedia('(prefers-color-scheme: dark)').matches ? Theme.Dark : Theme.Light
}

// The bootstrap-select library doesn't properly support the dark theme yet
// let's instead manually add some classes that make the selector look good
// when in dark mode.
export function setBootstrapSelectTheme(theme: Theme) {
    const searcherButton: HTMLElement | null = document.querySelector('.bootstrap-select button')
    if (searcherButton) {
        if (looksDark(theme)) {
            searcherButton.classList.remove('btn-light')
            searcherButton.classList.add('btn-outline-secondary', 'text-dark-emphasis')
        } else {
            searcherButton.classList.add('btn-light')
            searcherButton.classList.remove('btn-outline-secondary', 'text-dark-emphasis')
        }
    }
}

function looksDark(theme: Theme): Boolean {
    if (theme == Theme.Dark)
        return true

    if (theme != Theme.Light && window.matchMedia('(prefers-color-scheme: dark)').matches)
        return true

    return false
}

function setTheme(theme: Theme) {
    if (looksDark(theme))
        theme = Theme.Dark

    document.documentElement.setAttribute('data-bs-theme', theme)
}

function showActiveTheme(theme: Theme) {
    const themeSwitcher: HTMLElement = document.querySelector('#bd-theme')!
    const themeSwitcherText: HTMLElement = document.querySelector('#bd-theme-text')!
    const activeThemeIcon: HTMLElement = document.querySelector('.theme-icon-active use')!
    const btnToActive: HTMLElement = document.querySelector(`[data-bs-theme-value="${theme}"]`)!
    const svgOfActiveBtn: string = btnToActive.querySelector('svg use')!.getAttribute('href')!

    document.querySelectorAll('[data-bs-theme-value]').forEach(element => {
        element.classList.remove('active')
        element.setAttribute('aria-pressed', 'false')
    })

    btnToActive.classList.add('active')
    btnToActive.setAttribute('aria-pressed', 'true')
    activeThemeIcon.setAttribute('href', svgOfActiveBtn)
    const themeSwitcherLabel = `${themeSwitcherText.textContent} (${btnToActive.dataset.bsThemeValue})`
    themeSwitcher.setAttribute('aria-label', themeSwitcherLabel)

    document.querySelectorAll(`[data-theme-checkmark-for]:not([data-theme-checkmark-for="${theme}"])`)
        .forEach(element => element.classList.add('d-none'))

    document.querySelectorAll(`[data-theme-checkmark-for="${theme}"]`)
        .forEach(element => element.classList.remove('d-none'))
}

function focusThemeSwitcher() {
    const themeSwitcher: HTMLElement = document.querySelector('#bd-theme')!
    themeSwitcher.focus()
}

export function setupThemeHandling() {
    setTheme(getPreferredTheme())

    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        const storedTheme = getStoredTheme()
        if (storedTheme !== Theme.Light && storedTheme !== Theme.Dark) {
            setTheme(getPreferredTheme())
        }
    })

    window.addEventListener('DOMContentLoaded', () => {
        showActiveTheme(getPreferredTheme())

        document.querySelectorAll('[data-bs-theme-value]')
            .forEach(toggle => {
                toggle.addEventListener('click', () => {
                    const theme = toggle.getAttribute('data-bs-theme-value') as Theme
                    setStoredTheme(theme)
                    setTheme(theme)
                    showActiveTheme(theme)
                    setBootstrapSelectTheme(theme)
                    focusThemeSwitcher()
                })
            })
    })
}