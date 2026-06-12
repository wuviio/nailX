import 'package:dio/dio.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../../../core/api/client.dart';
import '../domain/design.dart';

part 'design_repository.g.dart';

@riverpod
DesignRepository designRepository(DesignRepositoryRef ref) {
  return DesignRepository(ref.watch(apiClientProvider));
}

@riverpod
Future<List<DesignIP>> designFeed(DesignFeedRef ref,
    {DesignFeedFilter? filter}) async {
  final repo = ref.watch(designRepositoryProvider);
  final f = filter ?? const DesignFeedFilter();
  final (designs, _) = await repo.listFeed(filter: f);
  return designs;
}

@riverpod
Future<DesignDetail> designDetail(DesignDetailRef ref, String id) async {
  return ref.watch(designRepositoryProvider).getDetail(id);
}

class DesignRepository {
  DesignRepository(this._dio);
  final Dio _dio;

  Future<(List<DesignIP>, String?)> listFeed({DesignFeedFilter? filter}) async {
    final f = filter ?? const DesignFeedFilter();
    final res = await _dio.get('/designs', queryParameters: {
      if (f.genderTag != null) 'gender_tag': f.genderTag,
      if (f.styleTag != null) 'style_tags': f.styleTag,
      'sort': f.sort,
      if (f.cursor != null) 'cursor': f.cursor,
      'limit': f.limit,
    });
    final designs = (res.data['designs'] as List)
        .map((e) => DesignIP.fromJson(e as Map<String, dynamic>))
        .toList();
    return (designs, res.data['next_cursor'] as String?);
  }

  Future<DesignDetail> getDetail(String id) async {
    final res = await _dio.get('/designs/$id');
    return DesignDetail.fromJson(res.data as Map<String, dynamic>);
  }

  Future<List<DesignIP>> listByUser(String userId, {String? cursor}) async {
    final res = await _dio.get('/users/$userId/designs', queryParameters: {
      if (cursor != null) 'cursor': cursor,
      'limit': 20,
    });
    return (res.data['designs'] as List)
        .map((e) => DesignIP.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<({String designIpId, String status, String jobId})> createDesign({
    required String title,
    String? description,
    required String previewImageUrl,
    required Map<String, dynamic> designData,
    String? parentIpId,
    required String genderTag,
    required List<String> styleTags,
    required bool isPublic,
  }) async {
    final res = await _dio.post('/designs', data: {
      'title': title,
      if (description != null) 'description': description,
      'preview_image_url': previewImageUrl,
      'design_data': designData,
      if (parentIpId != null) 'parent_ip_id': parentIpId,
      'gender_tag': genderTag,
      'style_tags': styleTags,
      'is_public': isPublic,
    });
    return (
      designIpId: res.data['design_ip_id'] as String,
      status: res.data['status'] as String,
      jobId: res.data['job_id'] as String,
    );
  }
}
