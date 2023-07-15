import { allLanguagesToplistPage, getMetadata, languageToplistPage } from './api.js';
import { Language, MetadataReponse } from './apitypes.js';
import { displayResultsLoadingSpinner, displayResults, displayResultsError, displayStatisticsLoadingSpinner, displayStatistics, displaySearcher } from './views.js';
import { goToAllLanguagesResults, goToOneLanguagesResults, routeFromHash } from './routes.js';
import { restoreScrollPosition, saveScrollPosition } from './scrollPosition.js';


async function resultsForAllLanguages(page: number, pages: number): Promise<void> {
    try {
        displayResultsLoadingSpinner()
        let allLanguagesPage = await allLanguagesToplistPage(page)

        displayResults(allLanguagesPage, page, pages, goToAllLanguagesResults)
    } catch(e: any) {
        displayResultsError(e)
    }
}

async function resultsForLanguage(language: Language, page: number): Promise<void> {
    try {
        displayResultsLoadingSpinner()
        let pageResults = await languageToplistPage(language.EscapedName, page)

        let onPageChange = (newPage: number) => {
            goToOneLanguagesResults(language, newPage)
        }
        displayResults(pageResults, page, language.Pages, onPageChange)
    } catch(e: any) {
        displayResultsError(e)
    }
}

async function routeResults(metadata: MetadataReponse): Promise<void> {
    await routeFromHash(metadata, {
        async resultsAllLanguages(page: number) {
            await resultsForAllLanguages(page, metadata.AllReposPages)
        },
        async resultsOneLanguage(language: Language, page: number) {
            await resultsForLanguage(language, page)
        },
    })
}

async function run(): Promise<void> {
    displayStatisticsLoadingSpinner()
    let metadata = await getMetadata()
    displayStatistics(metadata)

    await routeFromHash(metadata, {
        async resultsAllLanguages(_page: number) {
            displaySearcher(metadata, null)
        },
        async resultsOneLanguage(language: Language, _page: number) {
            displaySearcher(metadata, language)
        }
    })

    window.addEventListener('hashchange', () => {
        routeResults(metadata).catch(displayResultsError)
    })
    await routeResults(metadata)

    restoreScrollPosition()
    // save further scroll changes - but only start doing this after the previous
    // scroll position has been restored
    document.addEventListener('scrollend', saveScrollPosition)
}

run().catch(displayResultsError)