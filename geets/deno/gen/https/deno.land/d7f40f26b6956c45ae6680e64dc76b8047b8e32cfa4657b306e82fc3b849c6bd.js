import { fetchLocal, fetchRemote } from "../mod.ts";
export const fetchAuto = (path, onlyData)=>{
    try {
        new URL(path);
        return fetchRemote(path, onlyData);
    } catch (e) {
        return fetchLocal(path, onlyData);
    }
};
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC94L2ZldGNoYmFzZTY0QDEuMC4wL3NyYy9hdXRvLnRzPiJdLCJzb3VyY2VzQ29udGVudCI6WyJpbXBvcnQgeyBmZXRjaExvY2FsLCBmZXRjaFJlbW90ZSB9IGZyb20gXCIuLi9tb2QudHNcIlxuXG5leHBvcnQgY29uc3QgZmV0Y2hBdXRvID0gKHBhdGg6IHN0cmluZywgb25seURhdGE/OiBib29sZWFuKSA9PiB7XG4gICAgdHJ5IHtcbiAgICAgICAgbmV3IFVSTChwYXRoKVxuICAgICAgICByZXR1cm4gZmV0Y2hSZW1vdGUocGF0aCwgb25seURhdGEpXG4gICAgfSBjYXRjaChlKSB7XG4gICAgICAgIHJldHVybiBmZXRjaExvY2FsKHBhdGgsIG9ubHlEYXRhKVxuICAgIH1cbn0iXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IlNBQUEsVUFBQSxFQUFBLFdBQUEsU0FBQSxTQUFBO2FBRUEsU0FBQSxJQUFBLElBQUEsRUFBQSxRQUFBOztZQUVBLEdBQUEsQ0FBQSxJQUFBO2VBQ0EsV0FBQSxDQUFBLElBQUEsRUFBQSxRQUFBO2FBQ0EsQ0FBQTtlQUNBLFVBQUEsQ0FBQSxJQUFBLEVBQUEsUUFBQSJ9