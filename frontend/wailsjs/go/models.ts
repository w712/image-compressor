export namespace main {
	
	export class CompressOptions {
	    quality: number;
	    maxWidth: number;
	    maxHeight: number;
	    outputFormat: string;
	    outputDir: string;
	    keepAspect: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CompressOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.quality = source["quality"];
	        this.maxWidth = source["maxWidth"];
	        this.maxHeight = source["maxHeight"];
	        this.outputFormat = source["outputFormat"];
	        this.outputDir = source["outputDir"];
	        this.keepAspect = source["keepAspect"];
	    }
	}
	export class CompressResult {
	    success: boolean;
	    message: string;
	    originalSize: number;
	    newSize: number;
	    outputPath: string;
	    originalBase64: string;
	    compressedBase64: string;
	    originalWidth: number;
	    originalHeight: number;
	    newWidth: number;
	    newHeight: number;
	    compressionRatio: number;
	
	    static createFrom(source: any = {}) {
	        return new CompressResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.originalSize = source["originalSize"];
	        this.newSize = source["newSize"];
	        this.outputPath = source["outputPath"];
	        this.originalBase64 = source["originalBase64"];
	        this.compressedBase64 = source["compressedBase64"];
	        this.originalWidth = source["originalWidth"];
	        this.originalHeight = source["originalHeight"];
	        this.newWidth = source["newWidth"];
	        this.newHeight = source["newHeight"];
	        this.compressionRatio = source["compressionRatio"];
	    }
	}
	export class GifCompressOptions {
	    maxWidth: number;
	    maxHeight: number;
	    colors: number;
	    lossy: number;
	    outputDir: string;
	
	    static createFrom(source: any = {}) {
	        return new GifCompressOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxWidth = source["maxWidth"];
	        this.maxHeight = source["maxHeight"];
	        this.colors = source["colors"];
	        this.lossy = source["lossy"];
	        this.outputDir = source["outputDir"];
	    }
	}
	export class GifOptions {
	    frameDelay: number;
	    loopCount: number;
	    maxWidth: number;
	    maxHeight: number;
	    outputDir: string;
	    outputName: string;
	
	    static createFrom(source: any = {}) {
	        return new GifOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.frameDelay = source["frameDelay"];
	        this.loopCount = source["loopCount"];
	        this.maxWidth = source["maxWidth"];
	        this.maxHeight = source["maxHeight"];
	        this.outputDir = source["outputDir"];
	        this.outputName = source["outputName"];
	    }
	}
	export class GifResult {
	    success: boolean;
	    message: string;
	    outputPath: string;
	    fileSize: number;
	    frameCount: number;
	    width: number;
	    height: number;
	    preview: string;
	
	    static createFrom(source: any = {}) {
	        return new GifResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.outputPath = source["outputPath"];
	        this.fileSize = source["fileSize"];
	        this.frameCount = source["frameCount"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.preview = source["preview"];
	    }
	}
	export class ImageInfo {
	    path: string;
	    name: string;
	    size: number;
	    width: number;
	    height: number;
	    format: string;
	    preview: string;
	
	    static createFrom(source: any = {}) {
	        return new ImageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.format = source["format"];
	        this.preview = source["preview"];
	    }
	}

}

