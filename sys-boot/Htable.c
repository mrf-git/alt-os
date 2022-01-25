#include "Common.h"
#include "Htable.h"

//
// Hash table definitions.
//

#define MAX_HTABLE_BUCKET_KEYS      10
#define HTABLE_MODULUS              16001
#define HTABLE_SEED                 16007


//
// Structure definitions.
//

typedef struct {
    UINT16 BucketIndex;
    UINT16 KeyIndex;
} HTABLE_KEY_INDEX;

typedef struct {
    UINTN HtableKeyIndex;
    UINTN NumKeys;
    const CHAR8 *Keys[MAX_HTABLE_BUCKET_KEYS];
    void *Values[MAX_HTABLE_BUCKET_KEYS];
} HTABLE_BUCKET;

struct _SYS_HASH_TABLE {
    HTABLE_BUCKET Buckets[HTABLE_MODULUS];
    HTABLE_KEY_INDEX KeyIndexArray[HTABLE_MODULUS];
    UINTN NumKeys;
};


//
// Exported functions.
//

const UINTN SYSABI Sys_Htable_Size() {
    return sizeof(SYS_HASH_TABLE);
}


void SYSABI Sys_Htable_Init(IN OUT SYS_HASH_TABLE *Htable) {
    Htable->NumKeys = 0;
    for (UINTN i=0; i < HTABLE_MODULUS; i++) {
        Htable->Buckets[i].NumKeys = 0;
    }
}


BOOLEAN SYSABI Sys_Htable_Put(IN OUT SYS_HASH_TABLE *Htable, IN const CHAR8 *Key, IN void *Value) {
    UINTN BucketIndex = Sys_Common_StringHash(Key, HTABLE_SEED, HTABLE_MODULUS);
    HTABLE_BUCKET *Bucket = &Htable->Buckets[BucketIndex];

    if (Bucket->NumKeys > 0) {
        BOOLEAN Exists = FALSE;
        UINTN KeyIndex;
        for (UINTN i=0; i < Bucket->NumKeys; i++) {
            if (Sys_Common_AsciiStrCmp(Bucket->Keys[i], Key) == 0) {
                KeyIndex = i;
                Exists = TRUE;
                break;
            }
        }
        if (Exists) {
            // Update existing entry.
            Bucket->Keys[KeyIndex] = Key;
            Bucket->Values[KeyIndex] = Value;
            return TRUE;
        }
    }

    // Insert new entry.
    if (Bucket->NumKeys >= MAX_HTABLE_BUCKET_KEYS || Htable->NumKeys >= HTABLE_MODULUS) {
        return FALSE;
    }
    Bucket->Keys[Bucket->NumKeys] = Key;
    Bucket->Values[Bucket->NumKeys] = Value;
    Htable->KeyIndexArray[Htable->NumKeys].KeyIndex = Bucket->NumKeys;
    Bucket->NumKeys++;
    Htable->KeyIndexArray[Htable->NumKeys].BucketIndex = BucketIndex;
    Bucket->HtableKeyIndex = Htable->NumKeys;
    Htable->NumKeys++;
    return TRUE;
}


BOOLEAN SYSABI Sys_Htable_Get(IN OUT SYS_HASH_TABLE *Htable, IN const CHAR8 *Key, OUT void **Value) {
    UINTN BucketIndex = Sys_Common_StringHash(Key, HTABLE_SEED, HTABLE_MODULUS);
    HTABLE_BUCKET *Bucket = &Htable->Buckets[BucketIndex];
    if (Bucket->NumKeys == 0) {
        return FALSE;
    } else {
        for (UINTN i=0; i < Bucket->NumKeys; i++) {
            if (Sys_Common_AsciiStrCmp(Bucket->Keys[i], Key) == 0) {
                if (Value != NULL) {
                    *Value = Bucket->Values[i];
                }
                return TRUE;
            }
        }
    }
    return FALSE;
}


void SYSABI Sys_Htable_Remove(IN OUT SYS_HASH_TABLE *Htable, IN const CHAR8 *Key) {

    UINTN BucketIndex = Sys_Common_StringHash(Key, HTABLE_SEED, HTABLE_MODULUS);
    HTABLE_BUCKET *Bucket = &Htable->Buckets[BucketIndex];

    BOOLEAN Exists = FALSE;
    UINTN KeyIndex;
    for (UINTN i=0; i < Bucket->NumKeys; i++) {
        if (Sys_Common_AsciiStrCmp(Bucket->Keys[i], Key) == 0) {
            KeyIndex = i;
            Exists = TRUE;
            break;
        }
    }
    if (!Exists) {
        return;
    }

    // Shift any other keys in the bucket up.
    for (UINTN i=KeyIndex+1; i < Bucket->NumKeys; i++) {
        Bucket->Keys[i-1] = Bucket->Keys[i];
        Bucket->Values[i-1] = Bucket->Values[i];
        Htable->KeyIndexArray[Bucket->HtableKeyIndex].KeyIndex = i-1;
    }

    // Shift key array entries up.
    for (UINTN i=Bucket->HtableKeyIndex+1; i < Htable->NumKeys; i++) {
        Htable->KeyIndexArray[i-1].KeyIndex = Htable->KeyIndexArray[i].KeyIndex;
        Htable->KeyIndexArray[i-1].BucketIndex = Htable->KeyIndexArray[i].BucketIndex;
        HTABLE_BUCKET *ShiftedBucket = &Htable->Buckets[Htable->KeyIndexArray[i].BucketIndex];
        ShiftedBucket->HtableKeyIndex = i-1;
    }

    Bucket->NumKeys--;
    Htable->NumKeys--;
}


UINTN SYSABI Sys_Htable_NumKeys(IN OUT SYS_HASH_TABLE *Htable) {
    return Htable->NumKeys;
}


const CHAR8 * SYSABI Sys_Htable_Key(IN OUT SYS_HASH_TABLE *Htable, IN const UINTN Index) {
    if (Index >= Htable->NumKeys) {
        return NULL;
    }
    HTABLE_BUCKET *Bucket = &Htable->Buckets[Htable->KeyIndexArray[Index].BucketIndex];
    if (Htable->KeyIndexArray[Index].KeyIndex >= Bucket->NumKeys) {
        return NULL;
    }
    return Bucket->Keys[Htable->KeyIndexArray[Index].KeyIndex];
}

