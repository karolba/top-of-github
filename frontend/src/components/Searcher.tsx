import { h } from "dom-chef"
import { MetadataReponse } from "../apitypes"


export default function Searcher(metadata: MetadataReponse) {
    return (
        <div>
            <h3>Searcher</h3>
            <ul>
                {metadata.Languages.map(language => (
                    <li>{language.Name} - {language.CountOfRepos} repositories with {language.CountOfStars} stars</li>
                ))}
            </ul>
        </div>
    )
}