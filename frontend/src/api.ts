import { MetadataReponse, ToplistPageResponse } from './apitypes.js';

export async function getMetadata(): Promise<MetadataReponse> {
	let response = await fetch('https://top-of-github-api.baraniecki.eu/metadata.json.gz');
	return await response.json();
}

export async function languageToplistPage(escapedLanguageName: string, pageNumber: number): Promise<ToplistPageResponse> {
	let response = await fetch(`https://top-of-github-api.baraniecki.eu/language/${escapedLanguageName}/${pageNumber}.json.gz`);
	return await response.json();
}

export async function allLanguagesToplistPage(pageNumber: number): Promise<ToplistPageResponse> {
	let response = await fetch(`https://top-of-github-api.baraniecki.eu/all/${pageNumber}.json.gz`);
	return await response.json();
}
