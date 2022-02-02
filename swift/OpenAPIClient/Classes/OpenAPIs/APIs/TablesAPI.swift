//
// TablesAPI.swift
//
// Generated by openapi-generator
// https://openapi-generator.tech
//

import Foundation
#if canImport(AnyCodable)
import AnyCodable
#endif

open class TablesAPI {

    /**
     Create a new table
     
     - parameter table: (path) The name of the table to create 
     - parameter columns: (body) The array of column types to create for the new table. 
     - parameter apiResponseQueue: The queue on which api response is dispatched.
     - parameter completion: completion handler to receive the data and the error objects
     */
    open class func createTable(table: String, columns: ColumnCollection, apiResponseQueue: DispatchQueue = OpenAPIClientAPI.apiResponseQueue, completion: @escaping ((_ data: ColumnCollection?, _ error: Error?) -> Void)) {
        createTableWithRequestBuilder(table: table, columns: columns).execute(apiResponseQueue) { result in
            switch result {
            case let .success(response):
                completion(response.body, nil)
            case let .failure(error):
                completion(nil, error)
            }
        }
    }

    /**
     Create a new table
     - PUT /tables/{table}
     - parameter table: (path) The name of the table to create 
     - parameter columns: (body) The array of column types to create for the new table. 
     - returns: RequestBuilder<ColumnCollection> 
     */
    open class func createTableWithRequestBuilder(table: String, columns: ColumnCollection) -> RequestBuilder<ColumnCollection> {
        var localVariablePath = "/tables/{table}"
        let tablePreEscape = "\(APIHelper.mapValueToPathItem(table))"
        let tablePostEscape = tablePreEscape.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? ""
        localVariablePath = localVariablePath.replacingOccurrences(of: "{table}", with: tablePostEscape, options: .literal, range: nil)
        let localVariableURLString = OpenAPIClientAPI.basePath + localVariablePath
        let localVariableParameters = JSONEncodingHelper.encodingParameters(forEncodableObject: columns)

        let localVariableUrlComponents = URLComponents(string: localVariableURLString)

        let localVariableNillableHeaders: [String: Any?] = [
            :
        ]

        let localVariableHeaderParameters = APIHelper.rejectNilHeaders(localVariableNillableHeaders)

        let localVariableRequestBuilder: RequestBuilder<ColumnCollection>.Type = OpenAPIClientAPI.requestBuilderFactory.getBuilder()

        return localVariableRequestBuilder.init(method: "PUT", URLString: (localVariableUrlComponents?.string ?? localVariableURLString), parameters: localVariableParameters, headers: localVariableHeaderParameters)
    }

    /**
     List all tables for which the user has access.
     
     - parameter apiResponseQueue: The queue on which api response is dispatched.
     - parameter completion: completion handler to receive the data and the error objects
     */
    open class func listTables(apiResponseQueue: DispatchQueue = OpenAPIClientAPI.apiResponseQueue, completion: @escaping ((_ data: TableCollection?, _ error: Error?) -> Void)) {
        listTablesWithRequestBuilder().execute(apiResponseQueue) { result in
            switch result {
            case let .success(response):
                completion(response.body, nil)
            case let .failure(error):
                completion(nil, error)
            }
        }
    }

    /**
     List all tables for which the user has access.
     - GET /tables
     - returns: RequestBuilder<TableCollection> 
     */
    open class func listTablesWithRequestBuilder() -> RequestBuilder<TableCollection> {
        let localVariablePath = "/tables"
        let localVariableURLString = OpenAPIClientAPI.basePath + localVariablePath
        let localVariableParameters: [String: Any]? = nil

        let localVariableUrlComponents = URLComponents(string: localVariableURLString)

        let localVariableNillableHeaders: [String: Any?] = [
            :
        ]

        let localVariableHeaderParameters = APIHelper.rejectNilHeaders(localVariableNillableHeaders)

        let localVariableRequestBuilder: RequestBuilder<TableCollection>.Type = OpenAPIClientAPI.requestBuilderFactory.getBuilder()

        return localVariableRequestBuilder.init(method: "GET", URLString: (localVariableUrlComponents?.string ?? localVariableURLString), parameters: localVariableParameters, headers: localVariableHeaderParameters)
    }

    /**
     Get existing metadata information for columns in a table.
     
     - parameter table: (path) The name of the table to display 
     - parameter start: (query) Starting row of result set to return (optional)
     - parameter limit: (query) Limit on number of rows to return (optional)
     - parameter apiResponseQueue: The queue on which api response is dispatched.
     - parameter completion: completion handler to receive the data and the error objects
     */
    open class func showTable(table: String, start: Int? = nil, limit: Int? = nil, apiResponseQueue: DispatchQueue = OpenAPIClientAPI.apiResponseQueue, completion: @escaping ((_ data: ColumnCollection?, _ error: Error?) -> Void)) {
        showTableWithRequestBuilder(table: table, start: start, limit: limit).execute(apiResponseQueue) { result in
            switch result {
            case let .success(response):
                completion(response.body, nil)
            case let .failure(error):
                completion(nil, error)
            }
        }
    }

    /**
     Get existing metadata information for columns in a table.
     - GET /tables/{table}
     - parameter table: (path) The name of the table to display 
     - parameter start: (query) Starting row of result set to return (optional)
     - parameter limit: (query) Limit on number of rows to return (optional)
     - returns: RequestBuilder<ColumnCollection> 
     */
    open class func showTableWithRequestBuilder(table: String, start: Int? = nil, limit: Int? = nil) -> RequestBuilder<ColumnCollection> {
        var localVariablePath = "/tables/{table}"
        let tablePreEscape = "\(APIHelper.mapValueToPathItem(table))"
        let tablePostEscape = tablePreEscape.addingPercentEncoding(withAllowedCharacters: .urlPathAllowed) ?? ""
        localVariablePath = localVariablePath.replacingOccurrences(of: "{table}", with: tablePostEscape, options: .literal, range: nil)
        let localVariableURLString = OpenAPIClientAPI.basePath + localVariablePath
        let localVariableParameters: [String: Any]? = nil

        var localVariableUrlComponents = URLComponents(string: localVariableURLString)
        localVariableUrlComponents?.queryItems = APIHelper.mapValuesToQueryItems([
            "start": start?.encodeToJSON(),
            "limit": limit?.encodeToJSON(),
        ])

        let localVariableNillableHeaders: [String: Any?] = [
            :
        ]

        let localVariableHeaderParameters = APIHelper.rejectNilHeaders(localVariableNillableHeaders)

        let localVariableRequestBuilder: RequestBuilder<ColumnCollection>.Type = OpenAPIClientAPI.requestBuilderFactory.getBuilder()

        return localVariableRequestBuilder.init(method: "GET", URLString: (localVariableUrlComponents?.string ?? localVariableURLString), parameters: localVariableParameters, headers: localVariableHeaderParameters)
    }
}