import { MetadataReponse, Language } from './apitypes.js';
import { saveScrollPosition } from './scrollPosition.js';

function startsWithNumber(str: string): Boolean {
    return !!/^[0-9]/.test(str)
}

function isValidPageNumber(pageNumber: number, pages: number): Boolean {
    return pageNumber > 0 && pageNumber <= pages
}

export interface Routes {
    resultsAllLanguages: (page: number) => Promise<void>
    resultsOneLanguage: (language: Language, page: number) => Promise<void>
}

export async function routeFromHash(metadata: MetadataReponse, routes: Routes): Promise<void> {
    let hash = window.location.hash.replace(/^#/, '');

    if (startsWithNumber(hash) && isValidPageNumber(parseInt(hash, 10), metadata.AllReposPages)) {
        await routes.resultsAllLanguages(parseInt(hash, 10))
    } else {
        let [languageName, page] = hash.split('/');
        let language = metadata.Languages.find(language => language.EscapedName == languageName);
        if (language && page && isValidPageNumber(parseInt(page, 10), language.Pages)) {
            await routes.resultsOneLanguage(language, parseInt(page, 10))
        } else if (language) {
            await routes.resultsOneLanguage(language, 1)
        } else {
            await routes.resultsAllLanguages(1)
        }
    }
}

export function goToAllLanguagesResults(page: number) {
    saveScrollPosition()
    window.location.hash = `${page}`
}

export function goToOneLanguagesResults(language: Language, page: number) {
    saveScrollPosition()
    window.location.hash = `${language.EscapedName}/${page}`
}