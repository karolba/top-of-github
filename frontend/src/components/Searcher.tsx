import { h } from "dom-chef"
import { MetadataReponse } from "../apitypes"


export default function Searcher(metadata: MetadataReponse): JSX.Element {
    console.log("Searcher metadata", metadata)

    let languages = [
        {
            Name: "All languages",
            CountOfRepos: metadata.CountOfAllRepos,
            EscapedName: "",
        },
        ...metadata.Languages
    ]

    return (
        <>
            <div className="row m-3">
                <div className="col-md-3"></div>
                <div className="col-md-6">
                    <select id="searcher" className="selectpicker" data-live-search="true" data-width="100%">
                        {languages.map(language => 
                            <option data-subtext={` ${language.CountOfRepos} repos`}>
                                {language.Name ? language.Name : '(no language)'}
                            </option>
                        )}
                    </select>
                </div>
            </div>
        </>
    )
}