import { allLanguagesToplistPage, getMetadata, languageToplistPage, preloadAllLanguagesToplistPage, preloadLanguageToplistPage } from './api.js';
import { Language, MetadataReponse } from './apitypes.js';
import { displayResultsLoadingSpinner, displayResults, displayResultsError, displayStatisticsLoadingSpinner, displayStatistics, displaySearcher } from './views.js';
import { routePreloadFromHash, routeFromHash, allLanguagesResultsHref, oneLanguageResultsHref } from './routes.js';
import { restoreScrollPosition, saveScrollPosition } from './scrollPosition.js';
import { setupThemeHandling } from './colorModes.js';


async function resultsForAllLanguages(page: number, pages: number): Promise<void> {
    try {
        displayResultsLoadingSpinner()
        let allLanguagesPage = await allLanguagesToplistPage(page)

        displayResults(allLanguagesPage, page, pages, allLanguagesResultsHref)
    } catch(e: any) {
        displayResultsError(e)
    }
}

async function resultsForLanguage(language: Language, page: number): Promise<void> {
    try {
        displayResultsLoadingSpinner()
        let pageResults = await languageToplistPage(language.EscapedName, page)

        let newPageHref = (newPage: number): string => oneLanguageResultsHref(language, newPage)
        displayResults(pageResults, page, language.Pages, newPageHref)
    } catch(e: any) {
        displayResultsError(e)
    }
}

async function routeResults(metadata: MetadataReponse): Promise<void> {
    await routeFromHash(metadata, {
        async resultsAllLanguages(page: number) {
            displaySearcher(metadata, null)
            await resultsForAllLanguages(page, metadata.AllReposPages)
        },
        async resultsOneLanguage(language: Language, page: number) {
            displaySearcher(metadata, language)
            await resultsForLanguage(language, page)
        },
    })
}

function preloadApiResponse() {
    // Route without checking for correct language names and page counts, to have the results
    // be in flight before the "metadata" file appears
    routePreloadFromHash({
        async resultsAllLanguages(page: number) {
            preloadAllLanguagesToplistPage(page)
        },
        async resultsOneLanguage(language: Language, page: number) {
            preloadLanguageToplistPage(language.EscapedName, page)
        },
    })
}

async function run(): Promise<void> {
    setupThemeHandling()
    preloadApiResponse()
    displayStatisticsLoadingSpinner()
    let metadata = await getMetadata()
    displayStatistics(metadata)

    window.addEventListener('hashchange', async () => {
        await routeResults(metadata).catch(displayResultsError)
        restoreScrollPosition()
    })
    await routeResults(metadata)

    restoreScrollPosition()
    // save further scroll changes - but only start doing this after the previous
    // scroll position has been restored
    document.addEventListener('scrollend', saveScrollPosition)
}

run().catch(displayResultsError)