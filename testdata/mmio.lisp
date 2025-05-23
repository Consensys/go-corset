(module mmio)

(defcolumns
  (CN_A :i64 :display :dec)
  (CN_B :i64 :display :dec)
  (CN_C :i64 :display :dec)
  (INDEX_A :i64 :display :dec)
  (INDEX_B :i64 :display :dec)
  (INDEX_C :i64 :display :dec)
  (VAL_A :i128)
  (VAL_B :i128)
  (VAL_C :i128)
  (VAL_A_NEW :i128)
  (VAL_B_NEW :i128)
  (VAL_C_NEW :i128)
  (BYTE_A :byte@prove)
  (BYTE_B :byte@prove)
  (BYTE_C :byte@prove)
  (ACC_A :i128)
  (ACC_B :i128)
  (ACC_C :i128)
  (MMIO_STAMP :i32 :display :dec)
  (MMIO_INSTRUCTION :i16 :display :hex)
  (CONTEXT_SOURCE :i64 :display :dec)
  (CONTEXT_TARGET :i64 :display :dec)
  (SOURCE_LIMB_OFFSET :i64 :display :dec)
  (TARGET_LIMB_OFFSET :i64 :display :dec)
  (SOURCE_BYTE_OFFSET :i8 :display :dec)
  (TARGET_BYTE_OFFSET :i8 :display :dec)
  (SIZE :i64 :display :dec)
  (LIMB :i128)
  (TOTAL_SIZE :i64 :display :dec)
  (EXO_SUM :i32)
  (EXO_ID :i32)
  (KEC_ID :i32)
  (PHASE :i32)
  (SUCCESS_BIT :binary)
  (IS_LIMB_VANISHES :binary@prove)
  (IS_LIMB_TO_RAM_TRANSPLANT :binary@prove)
  (IS_LIMB_TO_RAM_ONE_TARGET :binary@prove)
  (IS_LIMB_TO_RAM_TWO_TARGET :binary@prove)
  (IS_RAM_TO_LIMB_TRANSPLANT :binary@prove)
  (IS_RAM_TO_LIMB_ONE_SOURCE :binary@prove)
  (IS_RAM_TO_LIMB_TWO_SOURCE :binary@prove)
  (IS_RAM_TO_RAM_TRANSPLANT :binary@prove)
  (IS_RAM_TO_RAM_PARTIAL :binary@prove)
  (IS_RAM_TO_RAM_TWO_TARGET :binary@prove)
  (IS_RAM_TO_RAM_TWO_SOURCE :binary@prove)
  (IS_RAM_EXCISION :binary@prove)
  (IS_RAM_VANISHES :binary@prove)
  (FAST :binary)
  (SLOW :binary)
  (EXO_IS_ROM :binary@prove)
  (EXO_IS_KEC :binary@prove)
  (EXO_IS_LOG :binary@prove)
  (EXO_IS_TXCD :binary@prove)
  (EXO_IS_ECDATA :binary@prove)
  (EXO_IS_RIPSHA :binary@prove)
  (EXO_IS_BLAKEMODEXP :binary@prove)
  (INDEX_X :i64)
  (BYTE_LIMB :byte@prove)
  (ACC_LIMB :i128)
  (BIT :binary :array [1:5])
  (ACC :i128 :array [1:4])
  (POW_256 :i128 :array [1:2])
  (COUNTER :i8))

(defalias
  STAMP MMIO_STAMP
  CT    COUNTER
  CNS   CONTEXT_SOURCE
  CNT   CONTEXT_TARGET
  SLO   SOURCE_LIMB_OFFSET
  SBO   SOURCE_BYTE_OFFSET
  TLO   TARGET_LIMB_OFFSET
  TBO   TARGET_BYTE_OFFSET)

(defconst
  LLARGE                                    16
  LLARGEMO                                  (- LLARGE 1)
  ;;
  EXO_SUM_INDEX_ROM                         0
  EXO_SUM_INDEX_KEC                         1
  EXO_SUM_INDEX_LOG                         2
  EXO_SUM_INDEX_TXCD                        3
  EXO_SUM_INDEX_ECDATA                      4
  EXO_SUM_INDEX_RIPSHA                      5
  EXO_SUM_INDEX_BLAKEMODEXP                 6
  EXO_SUM_WEIGHT_ROM                        (^ 2 EXO_SUM_INDEX_ROM)
  EXO_SUM_WEIGHT_KEC                        (^ 2 EXO_SUM_INDEX_KEC)
  EXO_SUM_WEIGHT_LOG                        (^ 2 EXO_SUM_INDEX_LOG)
  EXO_SUM_WEIGHT_TXCD                       (^ 2 EXO_SUM_INDEX_TXCD)
  EXO_SUM_WEIGHT_ECDATA                     (^ 2 EXO_SUM_INDEX_ECDATA)
  EXO_SUM_WEIGHT_RIPSHA                     (^ 2 EXO_SUM_INDEX_RIPSHA)
  EXO_SUM_WEIGHT_BLAKEMODEXP                (^ 2 EXO_SUM_INDEX_BLAKEMODEXP)
  ;;
  MMIO_INST_LIMB_VANISHES                   0xfe01
  MMIO_INST_LIMB_TO_RAM_TRANSPLANT          0xfe11
  MMIO_INST_LIMB_TO_RAM_ONE_TARGET          0xfe12
  MMIO_INST_LIMB_TO_RAM_TWO_TARGET          0xfe13
  MMIO_INST_RAM_TO_LIMB_TRANSPLANT          0xfe21
  MMIO_INST_RAM_TO_LIMB_ONE_SOURCE          0xfe22
  MMIO_INST_RAM_TO_LIMB_TWO_SOURCE          0xfe23
  MMIO_INST_RAM_TO_RAM_TRANSPLANT           0xfe31
  MMIO_INST_RAM_TO_RAM_PARTIAL              0xfe32
  MMIO_INST_RAM_TO_RAM_TWO_TARGET           0xfe33
  MMIO_INST_RAM_TO_RAM_TWO_SOURCE           0xfe34
  MMIO_INST_RAM_EXCISION                    0xfe41
  MMIO_INST_RAM_VANISHES                    0xfe42)

(module mmio)

(definterleaved
  CN_ABC
  (CN_A CN_B CN_C))

(definterleaved
  INDEX_ABC
  (INDEX_A INDEX_B INDEX_C))

(definterleaved
  MMIO_STAMP_3
  (MMIO_STAMP MMIO_STAMP MMIO_STAMP))

(definterleaved
  VAL_ABC
  (VAL_A VAL_B VAL_C))

(definterleaved
  VAL_ABC_NEW
  (VAL_A_NEW VAL_B_NEW VAL_C_NEW))

(defpermutation
  (CN_ABC_SORTED
   INDEX_ABC_SORTED
   MMIO_STAMP_3_SORTED
   VAL_ABC_SORTED
   VAL_ABC_NEW_SORTED)

  ((+ CN_ABC)
   (+ INDEX_ABC)
   (+ MMIO_STAMP_3)
   VAL_ABC
   VAL_ABC_NEW)
  )

(defconstraint memory-consistency (:guard CN_ABC_SORTED)
  (begin (if (will-remain-constant! CN_ABC_SORTED)
                (if (will-remain-constant! INDEX_ABC_SORTED)
                         (if-not (will-remain-constant! MMIO_STAMP_3_SORTED)
                                 (will-eq! VAL_ABC_SORTED VAL_ABC_NEW_SORTED))))
         (if-not (will-remain-constant! CN_ABC_SORTED)
                 (will-eq! VAL_ABC_SORTED 0))
         (if-not (will-remain-constant! INDEX_ABC_SORTED)
                 (will-eq! VAL_ABC_SORTED 0))))


(module mmio)

;;
;; Instruction decoding
;;
(defun (mmio_inst_weight)
  (+    (*    MMIO_INST_LIMB_VANISHES             IS_LIMB_VANISHES)
        (*    MMIO_INST_LIMB_TO_RAM_TRANSPLANT    IS_LIMB_TO_RAM_TRANSPLANT)
        (*    MMIO_INST_LIMB_TO_RAM_ONE_TARGET    IS_LIMB_TO_RAM_ONE_TARGET)
        (*    MMIO_INST_LIMB_TO_RAM_TWO_TARGET    IS_LIMB_TO_RAM_TWO_TARGET)
        (*    MMIO_INST_RAM_TO_LIMB_TRANSPLANT    IS_RAM_TO_LIMB_TRANSPLANT)
        (*    MMIO_INST_RAM_TO_LIMB_ONE_SOURCE    IS_RAM_TO_LIMB_ONE_SOURCE)
        (*    MMIO_INST_RAM_TO_LIMB_TWO_SOURCE    IS_RAM_TO_LIMB_TWO_SOURCE)
        (*    MMIO_INST_RAM_TO_RAM_TRANSPLANT     IS_RAM_TO_RAM_TRANSPLANT)
        (*    MMIO_INST_RAM_TO_RAM_PARTIAL        IS_RAM_TO_RAM_PARTIAL)
        (*    MMIO_INST_RAM_TO_RAM_TWO_TARGET     IS_RAM_TO_RAM_TWO_TARGET)
        (*    MMIO_INST_RAM_TO_RAM_TWO_SOURCE     IS_RAM_TO_RAM_TWO_SOURCE)
        (*    MMIO_INST_RAM_EXCISION              IS_RAM_EXCISION)
        (*    MMIO_INST_RAM_VANISHES              IS_RAM_VANISHES)))

(defconstraint decoding-mmio-instruction-flag ()
  (eq! MMIO_INSTRUCTION (mmio_inst_weight)))

(defun (fast-op-flag-sum)
  (+ IS_LIMB_VANISHES
     IS_LIMB_TO_RAM_TRANSPLANT
     IS_RAM_TO_LIMB_TRANSPLANT
     IS_RAM_TO_RAM_TRANSPLANT
     IS_RAM_VANISHES))

(defconstraint fast-op ()
  (eq! FAST (fast-op-flag-sum)))

(defun (slow-op-flag-sum)
  (+ IS_LIMB_TO_RAM_ONE_TARGET
     IS_LIMB_TO_RAM_TWO_TARGET
     IS_RAM_TO_LIMB_ONE_SOURCE
     IS_RAM_TO_LIMB_TWO_SOURCE
     IS_RAM_TO_RAM_PARTIAL
     IS_RAM_TO_RAM_TWO_TARGET
     IS_RAM_TO_RAM_TWO_SOURCE
     IS_RAM_EXCISION))

(defconstraint slow-op ()
  (eq! SLOW (slow-op-flag-sum)))

(defun (op-flag-sum)
  (+ (fast-op-flag-sum) (slow-op-flag-sum)))

(defconstraint no-stamp-no-op ()
  (eq! (op-flag-sum) (~ STAMP)))

(defun (weighted-exo-sum)
  (+ (* EXO_SUM_WEIGHT_ROM EXO_IS_ROM)
     (* EXO_SUM_WEIGHT_KEC EXO_IS_KEC)
     (* EXO_SUM_WEIGHT_LOG EXO_IS_LOG)
     (* EXO_SUM_WEIGHT_TXCD EXO_IS_TXCD)
     (* EXO_SUM_WEIGHT_ECDATA EXO_IS_ECDATA)
     (* EXO_SUM_WEIGHT_RIPSHA EXO_IS_RIPSHA)
     (* EXO_SUM_WEIGHT_BLAKEMODEXP EXO_IS_BLAKEMODEXP)))

(defconstraint exo-sum-decoding ()
  (eq! (weighted-exo-sum) (* (instruction-may-provide-exo-sum) EXO_SUM)))

(defun (instruction-may-provide-exo-sum)
  (+ IS_LIMB_VANISHES
     IS_LIMB_TO_RAM_TRANSPLANT
     IS_LIMB_TO_RAM_ONE_TARGET
     IS_LIMB_TO_RAM_TWO_TARGET
     IS_RAM_TO_LIMB_TRANSPLANT
     IS_RAM_TO_LIMB_ONE_SOURCE
     IS_RAM_TO_LIMB_TWO_SOURCE))

;;
;; Heartbeat
;;
(defconstraint first-row (:domain {0})
  (vanishes! MMIO_STAMP))

(defconstraint stamp-increment ()
  (or! (will-remain-constant! STAMP) (will-inc! STAMP 1)))

(defconstraint stamp-reset-ct ()
  (if-not-zero (- STAMP (prev STAMP))
               (vanishes! CT)))

(defconstraint fast-is-one-row (:guard FAST)
  (will-inc! STAMP 1))

(defconstraint slow-is-llarge-rows (:guard SLOW)
  (if-eq-else CT LLARGEMO (will-inc! STAMP 1) (will-inc! CT 1)))

(defconstraint last-row-is-finish (:domain {-1})
  (if-not-zero STAMP
               (if-eq SLOW 1 (eq! CT LLARGEMO))))

;;
;; Byte decompsition
;;
(defpurefun (byte-dec value byte acc ct)
  (begin (byte-decomposition ct acc byte)
         (if-eq ct LLARGEMO (eq! value acc))))

(defconstraint byte-decompositions ()
  (begin (byte-dec VAL_A BYTE_A ACC_A CT)
         (byte-dec VAL_B BYTE_B ACC_B CT)
         (byte-dec VAL_C BYTE_C ACC_C CT)
         (byte-dec LIMB BYTE_LIMB ACC_LIMB CT)))

;;
;; Counter constancies
;;
(defconstraint counter-constancies ()
  (begin (counter-constancy CT CN_A)
         (counter-constancy CT CN_B)
         (counter-constancy CT CN_C)
         (counter-constancy CT INDEX_A)
         (counter-constancy CT INDEX_B)
         (counter-constancy CT INDEX_C)
         (counter-constancy CT VAL_A)
         (counter-constancy CT VAL_B)
         (counter-constancy CT VAL_C)
         (counter-constancy CT VAL_A_NEW)
         (counter-constancy CT VAL_B_NEW)
         (counter-constancy CT VAL_C_NEW)
         (counter-constancy CT INDEX_X)))(module mmio)

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                                ;;
;;  MMIO instruction constraints  ;;
;;                                ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defconstraint limb-vanishes (:guard IS_LIMB_VANISHES)
               (begin (vanishes! CN_A)
                      (vanishes! CN_B)
                      (vanishes! CN_C)
                      (eq! INDEX_X TLO)
                      (vanishes! LIMB)))

(defconstraint limb-to-ram-transplant (:guard IS_LIMB_TO_RAM_TRANSPLANT)
               (begin (eq! CN_A CNT)
                      (vanishes! CN_B)
                      (vanishes! CN_B)
                      (eq! INDEX_A TLO)
                      (eq! VAL_A_NEW LIMB)
                      (eq! INDEX_X SLO)))

(defconstraint limb-to-ram-one-target (:guard IS_LIMB_TO_RAM_ONE_TARGET)
               (begin (eq! CN_A CNT)
                      (vanishes! CN_B)
                      (vanishes! CN_C)
                      (eq! INDEX_A TLO)
                      (eq! INDEX_X SLO)
                      (one-partial-to-one    VAL_A
                                             VAL_A_NEW
                                             BYTE_LIMB
                                             BYTE_A
                                             [ACC 1]
                                             [ACC 2]
                                             [POW_256 1]
                                             SBO
                                             TBO
                                             SIZE
                                             [BIT 1]
                                             [BIT 2]
                                             [BIT 3]
                                             [BIT 4]
                                             CT)))

(defconstraint limb-to-ram-two-target (:guard IS_LIMB_TO_RAM_TWO_TARGET)
               (begin (eq! CN_A CNT)
                      (eq! CN_B CNT)
                      (vanishes! CN_C)
                      (eq! INDEX_A TLO)
                      (eq! INDEX_B (+ TLO 1))
                      (eq! INDEX_X SLO)
                      (one-partial-to-two VAL_A
                                          VAL_B
                                          VAL_A_NEW
                                          VAL_B_NEW
                                          BYTE_LIMB
                                          BYTE_A
                                          BYTE_B
                                          [ACC 1]
                                          [ACC 2]
                                          [ACC 3]
                                          [ACC 4]
                                          [POW_256 1]
                                          SBO
                                          TBO
                                          SIZE
                                          [BIT 1]
                                          [BIT 2]
                                          [BIT 3]
                                          [BIT 4]
                                          [BIT 5]
                                          CT)))

(defconstraint ram-to-limb-transplant (:guard IS_RAM_TO_LIMB_TRANSPLANT)
               (begin (eq! CN_A CNS)
                      (vanishes! CN_B)
                      (vanishes! CN_C)
                      (eq! INDEX_A SLO)
                      (eq! VAL_A_NEW VAL_A)
                      (eq! INDEX_X TLO)
                      (eq! LIMB VAL_A)))

(defconstraint ram-to-limb-one-source (:guard IS_RAM_TO_LIMB_ONE_SOURCE)
               (begin (eq! CN_A CNS)
                      (vanishes! CN_B)
                      (vanishes! CN_C)
                      (eq! INDEX_A SLO)
                      (eq! VAL_A_NEW VAL_A)
                      (eq! INDEX_X TLO)
                      (one-to-one-padded    LIMB
                                            BYTE_A
                                            [ACC 1]
                                            [POW_256 1]
                                            SBO
                                            TBO
                                            SIZE
                                            [BIT 1]
                                            [BIT 2]
                                            [BIT 3]
                                            CT)))

(defconstraint ram-to-limb-two-source (:guard IS_RAM_TO_LIMB_TWO_SOURCE)
               (begin (eq! CN_A CNS)
                      (eq! CN_B CNS)
                      (vanishes! CN_C)
                      (eq! INDEX_A SLO)
                      (eq! INDEX_B (+ SLO 1))
                      (eq! VAL_A_NEW VAL_A)
                      (eq! VAL_B_NEW VAL_B)
                      (eq! INDEX_X TLO)
                      (two-to-one-padded LIMB
                                         BYTE_A
                                         BYTE_B
                                         [ACC 1]
                                         [ACC 2]
                                         [POW_256 1]
                                         [POW_256 2]
                                         SBO
                                         TBO
                                         SIZE
                                         [BIT 1]
                                         [BIT 2]
                                         [BIT 3]
                                         [BIT 4]
                                         CT)))

(defconstraint ram-to-ram-transplant (:guard IS_RAM_TO_RAM_TRANSPLANT)
               (begin (eq! CN_A CNS)
                      (eq! CN_B CNT)
                      (vanishes! CN_C)
                      (eq! INDEX_A SLO)
                      (eq! INDEX_B TLO)
                      (eq! VAL_A_NEW VAL_A)
                      (eq! VAL_B_NEW VAL_A)))

(defconstraint ram-to-ram-partial (:guard IS_RAM_TO_RAM_PARTIAL)
               (begin (eq! CN_A CNS)
                      (eq! CN_B CNT)
                      (vanishes! CN_C)
                      (eq! INDEX_A SLO)
                      (eq! INDEX_B TLO)
                      (eq! VAL_A_NEW VAL_A)
                      (one-partial-to-one    VAL_B
                                             VAL_B_NEW
                                             BYTE_A
                                             BYTE_B
                                             [ACC 1]
                                             [ACC 2]
                                             [POW_256 1]
                                             SBO
                                             TBO
                                             SIZE
                                             [BIT 1]
                                             [BIT 2]
                                             [BIT 3]
                                             [BIT 4]
                                             CT)))

(defconstraint ram-to-ram-two-target (:guard IS_RAM_TO_RAM_TWO_TARGET)
               (begin (eq! CN_A CNS)
                      (eq! CN_B CNT)
                      (eq! CN_C CNT)
                      (eq! INDEX_A SLO)
                      (eq! INDEX_B TLO)
                      (eq! INDEX_C (+ TLO 1))
                      (eq! VAL_A_NEW VAL_A)
                      (one-partial-to-two    VAL_B
                                             VAL_C
                                             VAL_B_NEW
                                             VAL_C_NEW
                                             BYTE_A
                                             BYTE_B
                                             BYTE_C
                                             [ACC 1]
                                             [ACC 2]
                                             [ACC 3]
                                             [ACC 4]
                                             [POW_256 1]
                                             SBO
                                             TBO
                                             SIZE
                                             [BIT 1]
                                             [BIT 2]
                                             [BIT 3]
                                             [BIT 4]
                                             [BIT 5]
                                             CT)))

(defconstraint ram-to-ram-two-source (:guard IS_RAM_TO_RAM_TWO_SOURCE)
               (begin (eq! CN_A CNS)
                      (eq! CN_B CNS)
                      (eq! CN_C CNT)
                      (eq! INDEX_A SLO)
                      (eq! INDEX_B (+ SLO 1))
                      (eq! INDEX_C TLO)
                      (eq! VAL_A_NEW VAL_A)
                      (eq! VAL_B_NEW VAL_B)
                      (two-partial-to-one    VAL_C
                                             VAL_C_NEW
                                             BYTE_A
                                             BYTE_B
                                             BYTE_C
                                             [ACC 1]
                                             [ACC 2]
                                             [ACC 3]
                                             [POW_256 1]
                                             [POW_256 2]
                                             SBO
                                             TBO
                                             SIZE
                                             [BIT 1]
                                             [BIT 2]
                                             [BIT 3]
                                             [BIT 4]
                                             CT)))

(defconstraint ram-excision (:guard IS_RAM_EXCISION)
               (begin (eq! CN_A CNT)
                      (vanishes! CN_B)
                      (vanishes! CN_C)
                      (eq! INDEX_A TLO)
                      (excision    VAL_A
                                   VAL_A_NEW
                                   BYTE_A
                                   [ACC 1]
                                   [POW_256 1]
                                   TBO
                                   SIZE
                                   [BIT 1]
                                   [BIT 2]
                                   CT)))

(defconstraint ram-vanishes (:guard IS_RAM_VANISHES)
               (begin (eq! CN_A CNT)
                      (vanishes! CN_B)
                      (vanishes! CN_C)
                      (eq! INDEX_A TLO)
                      (vanishes! VAL_A_NEW)))
(module mmio)

;;;;;;;;;;;;;;;;;;;;;;;;;
;;                     ;;
;;  Surgical Patterns  ;;
;;                     ;;
;;;;;;;;;;;;;;;;;;;;;;;;;

;; Excision
(defpurefun    (excision    target target_new
                            target_byte
                            accumulator
                            pow
                            target_marker
                            size
                            bit1
                            bit2
                            counter)
               (begin (plateau bit1 target_marker counter)
                      (plateau bit2 (+ target_marker size) counter)
                      (isolate-chunk accumulator target_byte bit1 bit2 counter)
                      (power pow bit2 counter)
                      (if-eq counter LLARGEMO
                             (eq! target_new
                                  (- target (* accumulator pow))))))

;; [1 => 1 Padded]
(defpurefun    (one-to-one-padded    target
                                     source_byte
                                     accumulator
                                     pow
                                     source_marker
                                     target_marker
                                     size
                                     bit1
                                     bit2
                                     bit3
                                     counter)
               (begin (plateau bit1 source_marker counter)
                      (plateau bit2 (+ source_marker size) counter)
                      (plateau bit3 (+ target_marker size) counter)
                      (isolate-chunk accumulator source_byte bit1 bit2 counter)
                      (power pow bit3 counter)
                      (if-eq counter LLARGEMO
                             (eq! target (* accumulator pow)))))

;; [2 => 1 Padded]
(defpurefun    (two-to-one-padded    target
                                     source1_byte
                                     source2_byte
                                     accumulator1
                                     accumulator2
                                     pow1
                                     pow2
                                     source1_marker
                                     target_marker
                                     size
                                     bit1
                                     bit2
                                     bit3
                                     bit4
                                     counter)
               (begin (plateau bit1
                               source1_marker
                               counter)
                      (plateau bit2
                               (+ source1_marker (- size LLARGE))
                               counter)
                      (plateau bit3
                               (- (+ target_marker LLARGE) source1_marker)
                               counter)
               (plateau bit4 (+ target_marker size) counter)
               (isolate-suffix accumulator1 source1_byte bit1 counter)
               (isolate-prefix accumulator2 source2_byte bit2 counter)
               (power pow1 bit3 counter)
               (power pow2 bit4 counter)
               (if-eq counter LLARGEMO
                      (eq! target
                           (+ (* accumulator1 pow1) (* accumulator2 pow2))))))

;; [1 Partial => 1]
(defpurefun (one-partial-to-one target
                                target_new
                                source_byte
                                target_byte
                                accumulator1
                                accumulator2
                                pow
                                source_marker
                                target_marker
                                size
                                bit1
                                bit2
                                bit3
                                bit4
                                counter)
            (begin (plateau bit1 target_marker counter)
                   (plateau bit2 (+ target_marker size) counter)
                   (plateau bit3 source_marker counter)
                   (plateau bit4 (+ source_marker size) counter)
                   (isolate-chunk accumulator1 target_byte bit1 bit2 counter)
                   (isolate-chunk accumulator2 source_byte bit3 bit4 counter)
                   (power pow bit2 counter)
                   (if-eq counter LLARGEMO
                          (eq! target_new
                               (+ target
                                  (* (- accumulator2 accumulator1) pow))))))

;; [1 Partial => 2]
(defpurefun (one-partial-to-two target1
                                target2
                                target1_new
                                target2_new
                                source_byte
                                target1_byte
                                target2_byte
                                accumulator1
                                accumulator2
                                accumulator3
                                accumulator4
                                pow
                                source_marker
                                target1_marker
                                size
                                bit1
                                bit2
                                bit3
                                bit4
                                bit5
                                counter)
            (begin (plateau bit1 target1_marker counter)
                   (plateau bit2
                            (- (+ target1_marker size) LLARGE)
                            counter)
                   (plateau bit3 source_marker counter)
                   (plateau bit4
                            (- (+ source_marker LLARGE) target1_marker)
                            counter)
                   (plateau bit5 (+ source_marker size) counter)
                   (isolate-suffix accumulator1 target1_byte bit1 counter)
                   (isolate-prefix accumulator2 target2_byte bit2 counter)
                   (isolate-chunk accumulator3 source_byte bit3 bit4 counter)
                   (isolate-chunk accumulator4 source_byte bit4 bit5 counter)
                   (power pow bit2 counter)
                   (if-eq counter LLARGEMO
                          (begin (eq! target1_new
                                      (+ target1 (- accumulator3 accumulator1)))
                                 (eq! target2_new
                                      (+ target2
                                         (* (- accumulator4 accumulator2) pow)))))))

;; [2 Partial => 1]
(defpurefun (two-partial-to-one target
                                target_new
                                source1_byte
                                source2_byte
                                target_byte
                                accumulator1
                                accumulator2
                                accumulator3
                                pow1
                                pow2
                                source_marker
                                target_marker
                                size
                                bit1
                                bit2
                                bit3
                                bit4
                                counter)
            (begin (plateau bit1 source_marker counter)
                   (plateau bit2
                            (- (+ source_marker size) LLARGE)
                            counter)
                   (plateau bit3 target_marker counter)
                   (plateau bit4 (+ target_marker size) counter)
                   (isolate-suffix accumulator1 source1_byte bit1 counter)
                   (isolate-prefix accumulator2 source2_byte bit2 counter)
                   (isolate-chunk accumulator3 target_byte bit3 bit4 counter)
                   (power pow1 bit4 counter)
                   (antipower pow2 bit2 counter)
                   (if-eq counter LLARGEMO
                            (eq! target_new
                                 (+ target
                                    (* (- (+ (* accumulator1 pow2) accumulator2)
                                          accumulator3)
                                       pow1))))))
(module mmio)

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                           ;;
;;  Specialized constraints  ;;
;;                           ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

;Plateau
(defpurefun (plateau x cst counter)
  (if-zero cst
           (eq! x 1)
           (if-zero counter
                    (vanishes! x)
                    (if-eq-else counter cst (eq! x 1) (remained-constant! x)))))

;Power
(defpurefun (power pow x counter)
  (if-zero counter
           (if-zero x
                    (eq! pow 1)
                    (eq! pow 256))
           (if-zero x
                    (remained-constant! pow)
                    (eq! pow
                         (* (prev pow) 256)))))

;AntiPower
(defpurefun (antipower pow x counter)
  (if-zero counter
           (if-zero x
                    (eq! pow 256)
                    (eq! pow 1))
           (if-zero x
                    (eq! pow
                         (* (prev pow) 256))
                    (remained-constant! pow))))

;IsolateSuffix
(defpurefun (isolate-suffix accumulator byte x counter)
  (if-zero counter
           (if-zero x
                    (vanishes! accumulator)
                    (eq! accumulator byte))
           (if-zero x
                    (remained-constant! accumulator)
                    (eq! accumulator
                         (+ (* 256 (prev accumulator))
                            byte)))))

;IsolatePrefix
(defpurefun (isolate-prefix accumulator byte x counter)
  (if-zero counter
           (if-zero x
                    (eq! accumulator byte)
                    (vanishes! accumulator))
           (if-zero x
                    (eq! accumulator
                         (+ (* 256 (prev accumulator))
                            byte))
                    (remained-constant! accumulator))))

;IsolateChunk
(defpurefun (isolate-chunk accumulator byte x y counter)
  (if-zero counter
           (if-zero x
                    (vanishes! accumulator)
                    (eq! accumulator byte))
           (if-zero x
                    (vanishes! accumulator)
                    (if-zero y
                             (eq! accumulator
                                  (+ (* 256 (prev accumulator))
                                     byte))
                             (remained-constant! accumulator)))))
