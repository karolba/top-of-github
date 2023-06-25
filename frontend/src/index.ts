import { MetadataReponse, ToplistPageResponse } from './apitypes.js';
import { allLanguagesToplistPage, getMetadata } from './api.js';
import Statistics from './components/Statistics.js';
import Searcher from './components/Searcher.js';
import Results from './components/Results.js';


function displayStatistics(metadata: MetadataReponse): void {
    document.getElementById('statistics-container')!.replaceChildren(Statistics(metadata))
}

function displaySearcher(metadata: MetadataReponse): void {
    document.getElementById('searcher-container')!.replaceChildren(Searcher(metadata))
}

function displayResults(languages: ToplistPageResponse): void {
    document.getElementById('searcher-container')!.replaceChildren(Results(languages))
}


async function run(): Promise<void> {
    let metadata = await getMetadata()
    displayStatistics(metadata)
    // displaySearcher(metadata)

    let allLanguagesPage1 = await allLanguagesToplistPage(1)
    displayResults(allLanguagesPage1)
}

run().catch(error => console.error("Caught an async error: ", error));
