import { MetadataReponse, ToplistPageResponse } from './apitypes.js';

// Use the dev-api-server when working locally
const BASE_API_URI = ['localhost', '127.0.0.1'].includes(document.location.hostname) 
    ? 'http://127.0.0.1:10002'
    : 'https://top-of-github-api.baraniecki.eu';

export async function getMetadata(): Promise<MetadataReponse> {
    let response = await fetch(`${BASE_API_URI}/metadata`);
    return await response.json();
}

export async function languageToplistPage(escapedLanguageName: string, pageNumber: number): Promise<ToplistPageResponse> {
    let response = await fetch(`${BASE_API_URI}/language/${escapedLanguageName}/${pageNumber}`);
    return await response.json();
}

export async function allLanguagesToplistPage(pageNumber: number): Promise<ToplistPageResponse> {
    let response = await fetch(`${BASE_API_URI}/all/${pageNumber}`);
    return await response.json();
}
