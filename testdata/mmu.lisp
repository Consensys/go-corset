(module mmu)

(defcolumns
  ;; shared columns
  (STAMP      :i32 :display :dec)
  (MMIO_STAMP :i32 :display :dec)
  ;; perspective selector
  (MACRO :binary@prove)
  (PRPRC :binary@prove)
  (MICRO :binary@prove)
  ;; OUTPUT OF THE PREPROCESSING
  (TOT :i32 :display :dec)
  (TOTLZ :i32 :display :dec)
  (TOTNT :i32 :display :dec)
  (TOTRZ :i32 :display :dec)
  (OUT :i64 :array [5])
  (BIN :binary :array [5])
  ;; MMU INSTRUCTION FLAG
  (IS_MLOAD                                :binary@prove)
  (IS_MSTORE                               :binary@prove)
  (IS_MSTORE8                              :binary@prove)
  (IS_INVALID_CODE_PREFIX                  :binary@prove)
  (IS_RIGHT_PADDED_WORD_EXTRACTION         :binary@prove)
  (IS_RAM_TO_EXO_WITH_PADDING              :binary@prove)
  (IS_EXO_TO_RAM_TRANSPLANTS               :binary@prove)
  (IS_RAM_TO_RAM_SANS_PADDING              :binary@prove)
  (IS_ANY_TO_RAM_WITH_PADDING_SOME_DATA    :binary@prove)
  (IS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING :binary@prove)
  (IS_MODEXP_ZERO                          :binary@prove)
  (IS_MODEXP_DATA                          :binary@prove)
  (IS_BLAKE                                :binary@prove)
  ;; USED ONLY IN MICRO ROW BUT ARE SHARED
  (LZRO     :binary@prove)
  (NT_ONLY  :binary@prove)
  (NT_FIRST :binary@prove)
  (NT_MDDL  :binary@prove)
  (NT_LAST  :binary@prove)
  (RZ_ONLY  :binary@prove)
  (RZ_FIRST :binary@prove)
  (RZ_MDDL  :binary@prove)
  (RZ_LAST  :binary@prove))

(defperspective macro

  ;; selector
  MACRO
  ((INST :i16 :display :hex)
   (SRC_ID :i64)
   (TGT_ID :i64)
   (AUX_ID :i64)
   (SRC_OFFSET_HI :i128)
   (SRC_OFFSET_LO :i128)
   (TGT_OFFSET_LO :i64)
   (SIZE          :i64)
   (REF_OFFSET    :i64)
   (REF_SIZE      :i64)
   (SUCCESS_BIT   :binary)
   (LIMB_1        :i128)
   (LIMB_2        :i128)
   (PHASE         :i32)
   (EXO_SUM       :i32)))

(defperspective prprc

  ;; selector
  PRPRC
  ((CT           :i16)
   (EUC_FLAG     :binary)
   (EUC_A        :i64)
   (EUC_B        :i64)
   (EUC_QUOT     :i64)
   (EUC_REM      :i64)
   (EUC_CEIL     :i64)
   (WCP_FLAG     :binary)
   (WCP_ARG_1_HI :i128)
   (WCP_ARG_1_LO :i128)
   (WCP_ARG_2_LO :i128)
   (WCP_RES      :binary)
   (WCP_INST     :byte :display :opcode)))

(defperspective micro

  ;; selector
  MICRO
  ((INST :i16 :display :hex)
   (SIZE :byte :display :dec)
   (SLO :i64 :display :dec)
   (SBO :byte :display :dec)
   (TLO :i64 :display :dec)
   (TBO :byte :display :dec)
   (LIMB :i128)
   (CN_S :i64 :display :dec)
   (CN_T :i64 :display :dec)
   (SUCCESS_BIT :binary)
   (EXO_SUM :i32)
   (PHASE :i32)
   (EXO_ID :i32)
   (KEC_ID :i32)
   (TOTAL_SIZE :i64 :display :dec)))


(module mmu)

(defconst
  ;;
  ;; MMU NB OF PP ROWS
  ;;
  NB_PP_ROWS_MLOAD                                   1
  NB_PP_ROWS_MSTORE                                  1
  NB_PP_ROWS_MSTORE8                                 1
  NB_PP_ROWS_INVALID_CODE_PREFIX                     1
  NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION            5
  NB_PP_ROWS_RAM_TO_EXO_WITH_PADDING                 4
  NB_PP_ROWS_EXO_TO_RAM_TRANSPLANTS                  1
  NB_PP_ROWS_RAM_TO_RAM_SANS_PADDING                 5
  NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING    4
  NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_SOME_DATA       10
  NB_PP_ROWS_MODEXP_ZERO                             1
  NB_PP_ROWS_MODEXP_DATA                             6
  NB_PP_ROWS_BLAKE                                   2
  ;;
  ;; MMU NB OF PP ROWS + 1
  ;;
  NB_PP_ROWS_MLOAD_PO                                (+ NB_PP_ROWS_MLOAD 1)
  NB_PP_ROWS_MSTORE_PO                               (+ NB_PP_ROWS_MSTORE 1)
  NB_PP_ROWS_MSTORE8_PO                              (+ NB_PP_ROWS_MSTORE8 1)
  NB_PP_ROWS_INVALID_CODE_PREFIX_PO                  (+ NB_PP_ROWS_INVALID_CODE_PREFIX 1)
  NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION_PO         (+ NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION 1)
  NB_PP_ROWS_RAM_TO_EXO_WITH_PADDING_PO              (+ NB_PP_ROWS_RAM_TO_EXO_WITH_PADDING 1)
  NB_PP_ROWS_EXO_TO_RAM_TRANSPLANTS_PO               (+ NB_PP_ROWS_EXO_TO_RAM_TRANSPLANTS 1)
  NB_PP_ROWS_RAM_TO_RAM_SANS_PADDING_PO              (+ NB_PP_ROWS_RAM_TO_RAM_SANS_PADDING 1)
  NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING_PO (+ NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING 1)
  NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_SOME_DATA_PO    (+ NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_SOME_DATA 1)
  NB_PP_ROWS_MODEXP_ZERO_PO                          (+ NB_PP_ROWS_MODEXP_ZERO 1)
  NB_PP_ROWS_MODEXP_DATA_PO                          (+ NB_PP_ROWS_MODEXP_DATA 1)
  NB_PP_ROWS_BLAKE_PO                                (+ NB_PP_ROWS_BLAKE 1)
  ;;
  ;; MMU NB OF PP ROWS + 2
  ;;
  NB_PP_ROWS_MLOAD_PT                                (+ NB_PP_ROWS_MLOAD 2)
  NB_PP_ROWS_MSTORE_PT                               (+ NB_PP_ROWS_MSTORE 2)
  NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION_PT         (+ NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION 2)
  NB_PP_ROWS_BLAKE_PT                                (+ NB_PP_ROWS_BLAKE 2)
  ;;
  ;; MMU NB OF micro-processing rows
  ;;
  NB_MICRO_ROWS_TOT_MLOAD                            2
  NB_MICRO_ROWS_TOT_MSTORE                           2
  NB_MICRO_ROWS_TOT_MSTORE_EIGHT                     1
  NB_MICRO_ROWS_TOT_INVALID_CODE_PREFIX              1
  NB_MICRO_ROWS_TOT_RIGHT_PADDED_WORD_EXTRACTION     2
  ;;NB_MICRO_ROWS_TOT_RAM_TO_EXO_WITH_PADDING              variable
  ;;NB_MICRO_ROWS_TOT_EXO_TO_RAM_TANSPLANTS                variable
  ;;NB_MICRO_ROWS_TOT_RAM_TO_RAM_SANS_PADDING              variable
  ;;NB_MICRO_ROWS_TOT_ANY_TO_RAM_WITH_PADDING_PURE_PADDING variable
  ;;NB_MICRO_ROWS_TOT_ANY_TO_RAM_WITH_PADDING_SOME_DATA    variable
  NB_MICRO_ROWS_TOT_MODEXP_ZERO                      32
  NB_MICRO_ROWS_TOT_MODEXP_DATA                      32
  NB_MICRO_ROWS_TOT_BLAKE                            2)


(module mmu)

(defun (flag-sum)
  (+ MACRO PRPRC MICRO))

(defconstraint perspective-flag ()
  (begin (debug (is-binary (flag-sum)))
         (if-zero STAMP
                  (vanishes! (flag-sum))
                  (eq! (flag-sum) 1))))

;;
;; Heartbeat
;;
(defconstraint first-row (:domain {0})
  (vanishes! STAMP)) ;; ""

(defconstraint no-macrostamp-no-microstamp ()
  (if-zero STAMP
           (vanishes! MMIO_STAMP)))

(defconstraint mmu-stamp-evolution ()
  (did-inc! STAMP MACRO))

(defconstraint mmio-stamp-evolution ()
  (did-inc! MMIO_STAMP MICRO))

(defconstraint prprc-after-macro (:guard MACRO)
  (eq! (next PRPRC) 1))

(defconstraint after-prprc (:guard PRPRC)
  (begin (debug (eq! (+ (next PRPRC) (next MICRO))
                     1))
         (if-zero prprc/CT
                  (will-eq! MICRO 1)
                  (begin (will-dec! prprc/CT 1)
                         (will-eq! PRPRC 1)))))

(defconstraint tot-nb-of-micro-inst ()
  (eq! TOT (+ TOTLZ TOTNT TOTRZ)))

(defconstraint after-micro (:guard MICRO)
  (begin (debug (eq! (+ (next MICRO) (next MACRO))
                     1))
         (did-dec! TOT 1)
         (if-zero TOT
                  (begin (will-eq! MACRO 1)
                         (debug (vanishes! TOTLZ))
                         (debug (vanishes! TOTNT))
                         (debug (vanishes! TOTRZ)))
                  (will-eq! MICRO 1))
         (if-zero (prev TOTLZ)
                  (vanishes! TOTLZ)
                  (did-dec! TOTLZ 1))
         (if-zero (prev TOTNT)
                  (vanishes! TOTNT)
                  (did-dec! (+ TOTLZ TOTNT) 1))))

(defconstraint last-row (:domain {-1}) ;; ""
  (if-not-zero STAMP
               (begin (eq! MICRO 1)
                      (vanishes! TOT))))

;;
;; Constancies
;;
(defun (prprc-constant X)
  (if-eq PRPRC 1 (remained-constant! X)))

(defconstraint prprc-constancies ()
  (begin (prprc-constant TOT)
         (debug (prprc-constant TOTLZ))
         (debug (prprc-constant TOTNT))
         (debug (prprc-constant TOTRZ))))

(defun (stamp-decrementing X)
  (if-not-zero (- STAMP
                  (+ (prev STAMP) 1))
               (any! (remained-constant! X) (did-dec! X 1))))

(defconstraint stamp-decrementings ()
  (begin (stamp-decrementing TOT)
         (stamp-decrementing TOTLZ)
         (stamp-decrementing TOTNT)
         (stamp-decrementing TOTRZ)))

(defun (stamp-constant X)
  (if-not-zero (- STAMP
                  (+ (prev STAMP) 1))
               (remained-constant! X)))

(defconstraint stamp-constancies ()
  (begin (for i [5] (stamp-constant [OUT i]))
         (for i [5] (stamp-constant [BIN i]))
         (stamp-constant (bin-flag-sum)))) ;; ""

(defun (micro-instruction-writing-constant X)
  (if-eq MICRO 1
         (if-eq (prev MICRO) 1 (remained-constant! X))))

(defconstraint mmio-row-constancies ()
  (begin (micro-instruction-writing-constant micro/CN_S)
         (micro-instruction-writing-constant micro/CN_T)
         (micro-instruction-writing-constant micro/SUCCESS_BIT)
         (micro-instruction-writing-constant micro/EXO_SUM)
         (micro-instruction-writing-constant micro/PHASE)
         (micro-instruction-writing-constant micro/EXO_ID)
         (micro-instruction-writing-constant micro/KEC_ID)
         (micro-instruction-writing-constant micro/TOTAL_SIZE)))

;;
;; Instruction Decoding
;;
(defun (bin-flag-sum)
  (+ (* 1 IS_MLOAD)
     (* 2 IS_MSTORE)
     (* 3 IS_MSTORE8)
     (* 4 IS_INVALID_CODE_PREFIX)
     (* 5 IS_RIGHT_PADDED_WORD_EXTRACTION)
     (* 6 IS_RAM_TO_EXO_WITH_PADDING)
     (* 7 IS_EXO_TO_RAM_TRANSPLANTS)
     (* 8 IS_RAM_TO_RAM_SANS_PADDING)
     (* 9 IS_ANY_TO_RAM_WITH_PADDING_SOME_DATA)
     (* 10 IS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING)
     (* 11 IS_MODEXP_ZERO)
     (* 12 IS_MODEXP_DATA)
     (* 13 IS_BLAKE)))

(defun (is-any-to-ram-with-padding)
  (force-bool (+    IS_ANY_TO_RAM_WITH_PADDING_SOME_DATA
                    IS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING)))

(defun (inst-flag-sum)
  (force-bool (+ IS_MLOAD
                 IS_MSTORE
                 IS_MSTORE8
                 IS_INVALID_CODE_PREFIX
                 IS_RIGHT_PADDED_WORD_EXTRACTION
                 IS_RAM_TO_EXO_WITH_PADDING
                 IS_EXO_TO_RAM_TRANSPLANTS
                 IS_RAM_TO_RAM_SANS_PADDING
                 (is-any-to-ram-with-padding)
                 IS_MODEXP_ZERO
                 IS_MODEXP_DATA
                 IS_BLAKE)))

(defun (weight-flag-sum)
  (+ (* MMU_INST_MLOAD IS_MLOAD)
     (* MMU_INST_MSTORE IS_MSTORE)
     (* MMU_INST_MSTORE8 IS_MSTORE8)
     (* MMU_INST_INVALID_CODE_PREFIX IS_INVALID_CODE_PREFIX)
     (* MMU_INST_RIGHT_PADDED_WORD_EXTRACTION IS_RIGHT_PADDED_WORD_EXTRACTION)
     (* MMU_INST_RAM_TO_EXO_WITH_PADDING IS_RAM_TO_EXO_WITH_PADDING)
     (* MMU_INST_EXO_TO_RAM_TRANSPLANTS IS_EXO_TO_RAM_TRANSPLANTS)
     (* MMU_INST_RAM_TO_RAM_SANS_PADDING IS_RAM_TO_RAM_SANS_PADDING)
     (* MMU_INST_ANY_TO_RAM_WITH_PADDING (is-any-to-ram-with-padding))
     (* MMU_INST_MODEXP_ZERO IS_MODEXP_ZERO)
     (* MMU_INST_MODEXP_DATA IS_MODEXP_DATA)
     (* MMU_INST_BLAKE IS_BLAKE)))

(defconstraint inst-flag-is-one ()
  (eq! (inst-flag-sum) (flag-sum)))

(defconstraint set-inst-flag (:guard MACRO)
  (eq! (weight-flag-sum) macro/INST))

;;
;; Micro Instruction writing row types
;;

(defun    (ntrv-row)    (force-bool (+ NT_ONLY NT_FIRST NT_MDDL NT_LAST)))
(defun    (rzro-row)    (force-bool (+ RZ_ONLY RZ_FIRST RZ_MDDL RZ_LAST)))
(defun    (zero-row)    (force-bool (+ LZRO (rzro-row))))

(defconstraint sum-row-flag ()
  (eq! (+ LZRO (ntrv-row) (rzro-row)) MICRO))

(defconstraint left-zero-decrements ()
  (if-eq LZRO 1 (did-dec! TOTLZ 1)))

(defconstraint nt-decrements ()
  (if-eq (ntrv-row) 1 (did-dec! TOTNT 1)))

(defconstraint right-zero-decrements ()
  (if-eq (rzro-row) 1 (did-dec! TOTRZ 1)))

(defconstraint is-nt-only-row (:guard NT_ONLY)
  (begin (vanishes! (prev (ntrv-row)))
         (vanishes! TOTNT)))

(defconstraint is-nt-first-row (:guard NT_FIRST)
  (begin (vanishes! (prev (ntrv-row)))
         (eq! (~ TOTNT) 1)))

(defconstraint is-nt-middle-row (:guard NT_MDDL)
  (begin (eq! (prev (ntrv-row)) 1)
         (eq! (~ TOTNT) 1)))

(defconstraint is-nt-last-row (:guard NT_LAST)
  (begin (eq! (prev (ntrv-row)) 1)
         (vanishes! TOTNT)))

(defconstraint is-rz-only-row (:guard RZ_ONLY)
  (begin (vanishes! (prev (rzro-row)))
         (vanishes! TOTRZ)))

(defconstraint is-rz-first-row (:guard RZ_FIRST)
  (begin (vanishes! (prev (rzro-row)))
         (eq! (~ TOTRZ) 1)))

(defconstraint is-rz-middle-row (:guard RZ_MDDL)
  (begin (eq! (prev (rzro-row)) 1)
         (eq! (~ TOTRZ) 1)))

(defconstraint is-rz-last-row (:guard RZ_LAST)
  (begin (eq! (prev (rzro-row)) 1)
         (vanishes! TOTRZ)))

;;
;; Setting nb of preprocessing rows
;;
(defconstraint set-prprc-ct-init (:guard MACRO)
  (eq! (next prprc/CT)
       (+ (* (- NB_PP_ROWS_MLOAD                                   1)    IS_MLOAD)
          (* (- NB_PP_ROWS_MSTORE                                  1)    IS_MSTORE)
          (* (- NB_PP_ROWS_MSTORE8                                 1)    IS_MSTORE8)
          (* (- NB_PP_ROWS_INVALID_CODE_PREFIX                     1)    IS_INVALID_CODE_PREFIX)
          (* (- NB_PP_ROWS_RIGHT_PADDED_WORD_EXTRACTION            1)    IS_RIGHT_PADDED_WORD_EXTRACTION)
          (* (- NB_PP_ROWS_RAM_TO_EXO_WITH_PADDING                 1)    IS_RAM_TO_EXO_WITH_PADDING)
          (* (- NB_PP_ROWS_EXO_TO_RAM_TRANSPLANTS                  1)    IS_EXO_TO_RAM_TRANSPLANTS)
          (* (- NB_PP_ROWS_RAM_TO_RAM_SANS_PADDING                 1)    IS_RAM_TO_RAM_SANS_PADDING)
          (* (- NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_SOME_DATA       1)    IS_ANY_TO_RAM_WITH_PADDING_SOME_DATA)
          (* (- NB_PP_ROWS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING    1)    IS_ANY_TO_RAM_WITH_PADDING_PURE_PADDING)
          (* (- NB_PP_ROWS_MODEXP_ZERO                             1)    IS_MODEXP_ZERO)
          (* (- NB_PP_ROWS_MODEXP_DATA                             1)    IS_MODEXP_DATA)
          (* (- NB_PP_ROWS_BLAKE                                   1)    IS_BLAKE))))

;;
;; Utilities
;;
(defun    (callToEuc    row_offset
                        dividend
                        divisor)
  (begin (eq! (shift prprc/EUC_FLAG    row_offset) 1)
         (eq! (shift prprc/EUC_A       row_offset) dividend)
         (eq! (shift prprc/EUC_B       row_offset) divisor)))

(defun    (callToLt     row_offset
                        arg1hi
                        arg1lo
                        arg2lo)
  (begin (eq! (shift prprc/WCP_FLAG        row_offset) 1)
         (eq! (shift prprc/WCP_INST        row_offset) EVM_INST_LT)
         (eq! (shift prprc/WCP_ARG_1_HI    row_offset) arg1hi)
         (eq! (shift prprc/WCP_ARG_1_LO    row_offset) arg1lo)
         (eq! (shift prprc/WCP_ARG_2_LO    row_offset) arg2lo)))

(defun    (callToEq     row_offset
                        arg1hi
                        arg1lo
                        arg2lo)
  (begin (eq! (shift prprc/WCP_FLAG        row_offset) 1)
         (eq! (shift prprc/WCP_INST        row_offset) EVM_INST_EQ)
         (eq! (shift prprc/WCP_ARG_1_HI    row_offset) arg1hi)
         (eq! (shift prprc/WCP_ARG_1_LO    row_offset) arg1lo)
         (eq! (shift prprc/WCP_ARG_2_LO    row_offset) arg2lo)))

(defun    (callToIszero row_offset
                        arg1hi
                        arg1lo)
  (begin (eq!              (shift prprc/WCP_FLAG        row_offset) 1)
         (eq!              (shift prprc/WCP_INST        row_offset) EVM_INST_ISZERO)
         (eq!              (shift prprc/WCP_ARG_1_HI    row_offset) arg1hi)
         (eq!              (shift prprc/WCP_ARG_1_LO    row_offset) arg1lo)
         (debug (vanishes! (shift prprc/WCP_ARG_2_LO    row_offset)))))

(defun    (standard-progression C)
  (eq! C
       (* (prev MICRO)
          (+ (prev C) 1))))
