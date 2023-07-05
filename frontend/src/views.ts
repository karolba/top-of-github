import { Language, MetadataReponse, ToplistPageResponse } from './apitypes.js';
import LoadingSpinner from './components/LoadingSpinner.js';
import Results from './components/Results.js';
import ResultsError from './components/ResultsError.js';
import Searcher from './components/Searcher.js';
import Statistics from './components/Statistics.js';
import { goToAllLanguagesResults, goToOneLanguagesResults } from './routes.js';

export function displayStatistics(metadata: MetadataReponse): void {
    document.getElementById('statistics-container')!.replaceChildren(Statistics(metadata));
}

export function displayStatisticsLoadingSpinner(): void {
    document.getElementById('statistics-container')!.replaceChildren(LoadingSpinner());
}

export function displaySearcher(metadata: MetadataReponse, language: Language | null): void {
    document.getElementById('searcher-container')!.replaceChildren(Searcher(metadata, language));

    // The bootstrap-select library is based on jQuery, so need to use a bit of jQuery here
    let jQuerySearcher = jQuery('select#searcher');
    jQuerySearcher.selectpicker({ showSubtext: true });
    jQuerySearcher.on('changed.bs.select', (_e, clickedIndex, _isSelected, _previousValue) => {
        if (clickedIndex == null) {
            clickedIndex = 0;
        }

        if (clickedIndex == 0) {
            goToAllLanguagesResults(1);
        } else {
            let selected = metadata.Languages[clickedIndex - 1];
            goToOneLanguagesResults(selected, 1);
        }
    });
}

export function displayResults(languages: ToplistPageResponse, page: number, pages: number, onPageChange: (page: number) => void): void {
    document.getElementById('results-container')!.replaceChildren(Results(languages, page, pages, onPageChange));
}

export function displayResultsLoadingSpinner(): void {
    document.getElementById('results-container')!.replaceChildren(LoadingSpinner());
}

export function displayResultsError(error: any): void {
    document.getElementById('results-container')!.replaceChildren(ResultsError(error));
}
