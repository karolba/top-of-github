import { Language, MetadataReponse, ToplistPageResponse } from './apitypes.js';
import { allLanguagesToplistPage, getMetadata, languageToplistPage } from './api.js';
import Statistics from './components/Statistics.js';
import Searcher from './components/Searcher.js';
import Results from './components/Results.js';
import LoadingSpinner from './components/LoadingSpinner.js';
import ResultsError from './components/ResultsError.js';

async function resultsForAllLanguages(page: number, pages: number): Promise<void> {
    window.location.hash = `${page}`

    try {
        displayResultsLoadingSpinner()
        let allLanguagesPage = await allLanguagesToplistPage(page)

        let onPageChange = (newPage: number) => {
            resultsForAllLanguages(newPage, pages)
        }
        displayResults(allLanguagesPage, page, pages, onPageChange)
    } catch(e: any) {
        displayResultsError(e)
    }
}
async function resultsForLanguage(language: Language, page: number): Promise<void> {
    window.location.hash = `${language.EscapedName}/${page}`

    try {
        displayResultsLoadingSpinner()
        let pageResults = await languageToplistPage(language.EscapedName, page)

        let onPageChange = (newPage: number) => {
            resultsForLanguage(language, newPage)
        }
        displayResults(pageResults, page, language.Pages, onPageChange)
    } catch(e: any) {
        displayResultsError(e)
    }
}

function displayStatistics(metadata: MetadataReponse): void {
    document.getElementById('statistics-container')!.replaceChildren(Statistics(metadata))
}

function displayStatisticsLoadingSpinner(): void {
    document.getElementById('statistics-container')!.replaceChildren(LoadingSpinner())
}

function displaySearcher(metadata: MetadataReponse, language: Language | null): void {
    document.getElementById('searcher-container')!.replaceChildren(Searcher(metadata, language))

    let jQuerySearcher = jQuery('select#searcher')
    jQuerySearcher.selectpicker({ showSubtext: true });
    jQuerySearcher.on('changed.bs.select', (_e, clickedIndex, _isSelected, _previousValue) => {
        if(clickedIndex == null) {
            clickedIndex = 0;
        }

        if(clickedIndex == 0) {
            resultsForAllLanguages(1, metadata.AllReposPages)
        } else {
            let selected = metadata.Languages[clickedIndex - 1]
            resultsForLanguage(selected, 1)
        }
    });
}

function displayResults(languages: ToplistPageResponse, page: number, pages: number, onPageChange: (page: number) => void): void {
    document.getElementById('results-container')!.replaceChildren(Results(languages, page, pages, onPageChange))
}

function displayResultsLoadingSpinner(): void {
    document.getElementById('results-container')!.replaceChildren(LoadingSpinner())
}

function displayResultsError(error: any): void {
    document.getElementById('results-container')!.replaceChildren(ResultsError(error))
}

function startsWithNumber(str: string): Boolean {
    return !!/^[0-9]/.test(str)
}

function isValidPageNumber(pageNumber: number, pages: number): Boolean {
    return pageNumber > 0 && pageNumber <= pages
}

async function run(): Promise<void> {
    displayStatisticsLoadingSpinner();
    let metadata = await getMetadata()
    displayStatistics(metadata)

    let hash = window.location.hash.replace(/^#/, '')

    if(startsWithNumber(hash) && isValidPageNumber(parseInt(hash, 10), metadata.AllReposPages)) {
        displaySearcher(metadata, null)
        resultsForAllLanguages(parseInt(hash, 10), metadata.AllReposPages)
    } else {
        let [languageName, page] = hash.split('/')
        let language = metadata.Languages.find(language => language.EscapedName == languageName)
        if(language && page && isValidPageNumber(parseInt(page, 10), language.Pages)) {
            displaySearcher(metadata, language)
            resultsForLanguage(language, parseInt(page, 10))
        } else if(language) {
            displaySearcher(metadata, language)
            resultsForLanguage(language, 1)
        } else {
            displaySearcher(metadata, null)
            resultsForAllLanguages(1,  metadata.AllReposPages)
        }
    }
}

run().catch(error => {
    console.error("Caught an async error: ", error)
    displayResultsError(error)
});