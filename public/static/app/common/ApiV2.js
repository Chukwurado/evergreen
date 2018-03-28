mciModule.factory('ApiV2', function(ApiUtil, API_V2) {
  var get = ApiUtil.httpGetter(API_V2.BASE)

  return {
    getPatchById: function(patchId) {
      return get(API_V2.PATCH_BY_ID, {patch_id: patchId})
    },
  }
})
