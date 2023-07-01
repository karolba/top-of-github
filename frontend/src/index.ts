import { Language, MetadataReponse, ToplistPageResponse } from './apitypes.js';
import { allLanguagesToplistPage, getMetadata, languageToplistPage } from './api.js';
import Statistics from './components/Statistics.js';
import Searcher from './components/Searcher.js';
import Results from './components/Results.js';
import LoadingSpinner from './components/LoadingSpinner.js';
import ResultsError from './components/ResultsError.js';

async function resultsForAllLanguages(page: number): Promise<void> {
    try {
        displayLoadingSpinner()
        let allLanguagesPage = await allLanguagesToplistPage(page)
        displayResults(allLanguagesPage)
    } catch(e: any) {
        displayResultsError(e)
    }
}
async function resultsForLanguage(language: Language, page: number): Promise<void> {
    try {
        displayLoadingSpinner()
        let pageResults = await languageToplistPage(language.EscapedName, page)
        displayResults(pageResults)
    } catch(e: any) {
        displayResultsError(e)
    }
}

function displayStatistics(metadata: MetadataReponse): void {
    document.getElementById('statistics-container')!.replaceChildren(Statistics(metadata))
}

function displaySearcher(metadata: MetadataReponse): void {
    document.getElementById('searcher-container')!.replaceChildren(Searcher(metadata))

    let jQuerySearcher = jQuery('select#searcher')
    jQuerySearcher.selectpicker({ showSubtext: true });
    jQuerySearcher.on('changed.bs.select', (_e, clickedIndex, _isSelected, _previousValue) => {
        if(clickedIndex == null) {
            clickedIndex = 0;
        }

        if(clickedIndex == 0) {
            resultsForAllLanguages(1)
        } else {
            let selected = metadata.Languages[clickedIndex - 1]
            resultsForLanguage(selected, 1)
        }
    });
}

function displayResults(languages: ToplistPageResponse): void {
    document.getElementById('results-container')!.replaceChildren(Results(languages))
}

function displayLoadingSpinner(): void {
    document.getElementById('results-container')!.replaceChildren(LoadingSpinner())
}

function displayResultsError(error: any): void {
    document.getElementById('results-container')!.replaceChildren(ResultsError(error))
}

async function run(): Promise<void> {
    let metadata = await getMetadata()
    displayStatistics(metadata)
    displaySearcher(metadata)
    resultsForAllLanguages(1)
}

run().catch(error => console.error("Caught an async error: ", error));
