// SPDX-License-Identifier: GPL-3.0-or-later

#import <Cocoa/Cocoa.h>
#import <string.h>

// Returns every app path registered to open https:// URLs, newline-joined,
// or NULL if there are none. Caller must free() the result.
char *browserscan_nsworkspace_app_paths(void) {
    @autoreleasepool {
        NSURL *probe = [NSURL URLWithString:@"https://example.com"];
        NSArray<NSURL *> *appURLs =
            [[NSWorkspace sharedWorkspace] URLsForApplicationsToOpenURL:probe];

        if (appURLs.count == 0) {
            return NULL;
        }

        NSMutableArray<NSString *> *paths = [NSMutableArray array];
        for (NSURL *url in appURLs) {
            [paths addObject:url.path];
        }
        NSString *joined = [paths componentsJoinedByString:@"\n"];
        // Copy out of the autorelease pool so the caller owns the returned memory.
        return strdup(joined.UTF8String);
    }
}
