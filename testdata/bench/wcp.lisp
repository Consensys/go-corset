(module wcp)

(defcolumns
  (WORD_COMPARISON_STAMP :i32)
  (COUNTER :byte)
  (CT_MAX :byte)
  (INST :byte :display :opcode)
  (ARGUMENT_1_HI :i128)
  (ARGUMENT_1_LO :i128)
  (ARGUMENT_2_HI :i128)
  (ARGUMENT_2_LO :i128)
  (RESULT :binary@prove)
  (IS_LT :binary@prove)
  (IS_GT :binary@prove)
  (IS_SLT :binary@prove)
  (IS_SGT :binary@prove)
  (IS_EQ :binary@prove)
  (IS_ISZERO :binary@prove)
  (IS_GEQ :binary@prove)
  (IS_LEQ :binary@prove)
  (ONE_LINE_INSTRUCTION :binary)
  (VARIABLE_LENGTH_INSTRUCTION :binary)
  (BITS :binary@prove)
  (NEG_1 :binary@prove)
  (NEG_2 :binary@prove)
  (BYTE_1 :byte@prove)
  (BYTE_2 :byte@prove)
  (BYTE_3 :byte@prove)
  (BYTE_4 :byte@prove)
  (BYTE_5 :byte@prove)
  (BYTE_6 :byte@prove)
  (ACC_1 :i128)
  (ACC_2 :i128)
  (ACC_3 :i128)
  (ACC_4 :i128)
  (ACC_5 :i128)
  (ACC_6 :i128)
  (BIT_1 :binary@prove)
  (BIT_2 :binary@prove)
  (BIT_3 :binary@prove)
  (BIT_4 :binary@prove))

;; aliases
(defalias
  STAMP    WORD_COMPARISON_STAMP
  OLI      ONE_LINE_INSTRUCTION
  VLI      VARIABLE_LENGTH_INSTRUCTION
  CT       COUNTER
  ARG_1_HI ARGUMENT_1_HI
  ARG_1_LO ARGUMENT_1_LO
  ARG_2_HI ARGUMENT_2_HI
  ARG_2_LO ARGUMENT_2_LO
  RES      RESULT)

;; opcode values
(defconst
  EVM_INST_LT                               0x10
  EVM_INST_GT                               0x11
  EVM_INST_SLT                              0x12
  EVM_INST_SGT                              0x13
  EVM_INST_EQ                               0x14
  EVM_INST_ISZERO                           0x15
  ;;
  WCP_INST_GEQ                              0x0E
  WCP_INST_LEQ                              0x0F
  ;;
  LLARGE                                    16
  LLARGEMO                                  (- LLARGE 1))

(module wcp)

(defun (flag-sum)
  (+ (one-line-inst) (variable-length-inst)))

(defun (weight-sum)
  (+
    (* EVM_INST_LT     IS_LT)
    (* EVM_INST_GT     IS_GT)
    (* EVM_INST_SLT    IS_SLT)
    (* EVM_INST_SGT    IS_SGT)
    (* EVM_INST_EQ     IS_EQ)
    (* EVM_INST_ISZERO IS_ISZERO)
    (* WCP_INST_GEQ    IS_GEQ)
    (* WCP_INST_LEQ    IS_LEQ)))

(defun (one-line-inst)
  (+ IS_EQ IS_ISZERO))

(defun (variable-length-inst)
  (+ IS_LT IS_GT IS_LEQ IS_GEQ IS_SLT IS_SGT))

(defconstraint inst-decoding ()
  (if (== STAMP 0)
      (== (flag-sum) 0)
      (== (flag-sum) 1)))

(defconstraint setting-flag ()
  (begin
   (== INST (weight-sum))
   (== OLI (one-line-inst))
   (== VLI (variable-length-inst))))

(defconstraint counter-constancies ()
  (begin
   (counter-constancy CT ARG_1_HI)
   (counter-constancy CT ARG_1_LO)
   (counter-constancy CT ARG_2_HI)
   (counter-constancy CT ARG_2_LO)
   (counter-constancy CT RES)
   (counter-constancy CT INST)
   (counter-constancy CT CT_MAX)
   (counter-constancy CT BIT_3)
   (counter-constancy CT BIT_4)
   (counter-constancy CT NEG_1)
   (counter-constancy CT NEG_2)))

(defconstraint first-row (:domain {0})
  (== STAMP 0))

(defconstraint stamp-increments ()
  (∨ (will-remain-constant! STAMP) (will-inc! STAMP 1)))

(defconstraint counter-reset ()
  (if (¬ (will-remain-constant! STAMP))
      (== (next CT) 0)))

(defconstraint setting-ct-max ()
  (if (== OLI 1)
      (== CT_MAX 0)))

(defconstraint heartbeat (:guard STAMP)
  (if (== CT CT_MAX)
      (will-inc! STAMP 1)
      (will-inc! CT 1)))

(defconstraint ct-upper-bond ()
  (!= LLARGE CT))

(defconstraint lastRow (:domain {-1})
  (== CT CT_MAX))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                              ;;
;;    2.6 byte decompositions   ;;
;;                              ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; byte decompositions
(defconstraint byte_decompositions ()
  (begin (byte-decomposition CT ACC_1 BYTE_1)
         (byte-decomposition CT ACC_2 BYTE_2)
         (byte-decomposition CT ACC_3 BYTE_3)
         (byte-decomposition CT ACC_4 BYTE_4)
         (byte-decomposition CT ACC_5 BYTE_5)
         (byte-decomposition CT ACC_6 BYTE_6)))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                                         ;;
;;    2.7 BITS and sign bit constraints    ;;
;;                                         ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
(defconstraint bits-and-negs (:guard (+ IS_SLT IS_SGT))
  (if (== CT LLARGEMO)
         (begin (== (shift BYTE_1 (- 0 LLARGEMO))
                     (first-eight-bits-bit-dec))
                (== (shift BYTE_3 (- 0 LLARGEMO))
                     (last-eight-bits-bit-dec))
                (== NEG_1
                     (shift BITS (- 0 LLARGEMO)))
                (== NEG_2
                     (shift BITS (- 0 7))))))

(defconstraint no-neg-if-small ()
  (if (!= CT_MAX LLARGEMO)
      (begin (== NEG_1 0)
             (== NEG_2 0))))

(defun (first-eight-bits-bit-dec)
  (reduce +
          (for i
               [0 :7]
               (* (^ 2 i)
                  (shift BITS
                         (- 0 (+ i 8)))))))

(defun (last-eight-bits-bit-dec)
  (reduce +
          (for i
               [0 :7]
               (* (^ 2 i)
                  (shift BITS (- 0 i))))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                              ;;
;;    2.6 target constraints    ;;
;;                              ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
(defconstraint target-constraints ()
  (begin
   (if (!= STAMP 0)
       (begin
        (if (== ARG_1_HI ARG_2_HI)
            (== BIT_1 1)
            (== BIT_1 0))
        (if (== ARG_1_LO ARG_2_LO)
            (== BIT_2 1)
            (== BIT_2 0))))
   (if (== VLI 1)
       (if (== CT CT_MAX)
           (begin
            (== ACC_1 ARG_1_HI)
            (== ACC_2 ARG_1_LO)
            (== ACC_3 ARG_2_HI)
            (== ACC_4 ARG_2_LO)
            (== ACC_5
                (- (* (- (* 2 BIT_3) 1)
                      (- ARG_1_HI ARG_2_HI))
                   BIT_3))
            (== ACC_6
                (- (* (- (* 2 BIT_4) 1)
                      (- ARG_1_LO ARG_2_LO))
                   BIT_4)))))
   (if (== IS_ISZERO 1)
       (begin
        (== ARG_2_HI 0)
        (== ARG_2_LO 0)))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                              ;;
;;    2.7 result constraints    ;;
;;                              ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; eq_ = [[1]] . [[2]]
;; gt_ = [[3]] + [[1]] . [[4]]
;; lt_ = 1 - eq - gt
(defun (eq_)
  (* BIT_1 BIT_2))

(defun (gt_)
  (+ BIT_3 (* BIT_1 BIT_4)))

(defun (lt_)
  (- 1 (eq_) (gt_)))

;; 2.7.2
(defconstraint result ()
  (begin
   (if (== OLI 1) (== RES (eq_)))
   (if (== IS_LT 1) (== RES (lt_)))
   (if (== IS_GT 1) (== RES (gt_)))
   (if (== IS_LEQ 1)
       (== RES (+ (lt_) (eq_))))
   (if (== IS_GEQ 1)
       (== RES (+ (gt_) (eq_))))
   (if (== IS_SLT 1)
       (if (== NEG_1 NEG_2)
           (== RES (lt_))
           (== RES NEG_1)))
   (if (== IS_SGT 1)
       (if (== NEG_1 NEG_2)
           (== RES (gt_))
           (== RES NEG_2)))))
