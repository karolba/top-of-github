import { MetadataReponse, ToplistPageResponse } from './apitypes.js';
import { allLanguagesToplistPage, getMetadata } from './api.js';
import Statistics from './components/Statistics.js';
import Searcher from './components/Searcher.js';
import Results from './components/Results.js';


function displayStatistics(metadata: MetadataReponse) {
	document.getElementById('statistics-container')!.replaceChildren(Statistics(metadata))
}

function displaySearcher(metadata: MetadataReponse) {
	document.getElementById('searcher-container')!.replaceChildren(Searcher(metadata))
}

function displayResults(languages: ToplistPageResponse) {
	document.getElementById('searcher-container')!.replaceChildren(Results(languages))
}


async function run() {
	let metadata = await getMetadata()
	displayStatistics(metadata)
	displaySearcher(metadata)

	let allLanguagesPage1 = await allLanguagesToplistPage(1)
	displayResults(allLanguagesPage1)
}

run().catch(error => console.error("Caught an async error: ", error));

		// <div className="my-3 p-3 bg-body rounded shadow-sm">
		// 	<h6 className="border-bottom pb-2 mb-0">Suggestions</h6>
		// 	<div className="d-flex text-body-secondary pt-3">
		// 		<svg className="bd-placeholder-img flex-shrink-0 me-2 rounded" width="32" height="32"
		// 			xmlns="http://www.w3.org/2000/svg" role="img" aria-label="Placeholder: 32x32"
		// 			preserveAspectRatio="xMidYMid slice" focusable="false">
		// 			<title>Placeholder</title>
		// 			<rect width="100%" height="100%" fill="#007bff" /><text x="50%" y="50%" fill="#007bff"
		// 				dy=".3em">32x32</text>
		// 		</svg>
		// 		<div className="pb-3 mb-0 small lh-sm border-bottom w-100">
		// 			<div className="d-flex justify-content-between">
		// 				<strong className="text-gray-dark">Full Name</strong>
		// 				<a href="#">Follow</a>
		// 			</div>
		// 			<span className="d-block">@username</span>
		// 		</div>
		// 	</div>
		// 	<div className="d-flex text-body-secondary pt-3">
		// 		<svg className="bd-placeholder-img flex-shrink-0 me-2 rounded" width="32" height="32"
		// 			xmlns="http://www.w3.org/2000/svg" role="img" aria-label="Placeholder: 32x32"
		// 			preserveAspectRatio="xMidYMid slice" focusable="false">
		// 			<title>Placeholder</title>
		// 			<rect width="100%" height="100%" fill="#007bff" /><text x="50%" y="50%" fill="#007bff"
		// 				dy=".3em">32x32</text>
		// 		</svg>
		// 		<div className="pb-3 mb-0 small lh-sm border-bottom w-100">
		// 			<div className="d-flex justify-content-between">
		// 				<strong className="text-gray-dark">Full Name</strong>
		// 				<a href="#">Follow</a>
		// 			</div>
		// 			<span className="d-block">@username</span>
		// 		</div>
		// 	</div>
		// 	<div className="d-flex text-body-secondary pt-3">
		// 		<svg className="bd-placeholder-img flex-shrink-0 me-2 rounded" width="32" height="32"
		// 			xmlns="http://www.w3.org/2000/svg" role="img" aria-label="Placeholder: 32x32"
		// 			preserveAspectRatio="xMidYMid slice" focusable="false">
		// 			<title>Placeholder</title>
		// 			<rect width="100%" height="100%" fill="#007bff" /><text x="50%" y="50%" fill="#007bff"
		// 				dy=".3em">32x32</text>
		// 		</svg>
		// 		<div className="pb-3 mb-0 small lh-sm border-bottom w-100">
		// 			<div className="d-flex justify-content-between">
		// 				<strong className="text-gray-dark">Full Name</strong>
		// 				<a href="#">Follow</a>
		// 			</div>
		// 			<span className="d-block">@username</span>
		// 		</div>
		// 	</div>
		// 	<small className="d-block text-end mt-3">
		// 		<a href="#">All suggestions</a>
		// 	</small>
		// </div>