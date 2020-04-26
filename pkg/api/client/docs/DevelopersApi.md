# \DevelopersApi

All URIs are relative to *https://virtserver.swaggerhub.com/sebasmannem/guacamole/1.0.0*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SearchInventory**](DevelopersApi.md#SearchInventory) | **Get** /inventory | searches inventory


# **SearchInventory**
> []InventoryItem SearchInventory(ctx, optional)
searches inventory

By passing in the appropriate options, you can search for available inventory in the system 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***DevelopersApiSearchInventoryOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DevelopersApiSearchInventoryOpts struct

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **searchString** | **optional.String**| pass an optional search string for looking up inventory | 
 **skip** | **optional.Int32**| number of records to skip for pagination | 
 **limit** | **optional.Int32**| maximum number of records to return | 

### Return type

[**[]InventoryItem**](InventoryItem.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

