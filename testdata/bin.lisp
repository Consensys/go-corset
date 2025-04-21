(module bin)

(defcolumns
  (STAMP :i32)
  (CT_MAX :byte)
  (COUNTER :byte)
  (INST :byte :display :opcode)
  (ARGUMENT_1_HI :i128)
  (ARGUMENT_1_LO :i128)
  (ARGUMENT_2_HI :i128)
  (ARGUMENT_2_LO :i128)
  (RESULT_HI :i128)
  (RESULT_LO :i128)
  (IS_AND :binary@prove)
  (IS_OR :binary@prove)
  (IS_XOR :binary@prove)
  (IS_NOT :binary@prove)
  (IS_BYTE :binary@prove)
  (IS_SIGNEXTEND :binary@prove)
  (SMALL :binary@prove)
  (BITS :binary@prove)
  (BIT_B_4 :binary@prove)
  (LOW_4 :byte)
  (NEG :binary@prove)
  (BIT_1 :binary@prove)
  (PIVOT :byte)
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
  ;; decoded bytes:
  (XXX_BYTE_HI :byte)
  (XXX_BYTE_LO :byte))

;; aliases
(defalias
  CT       COUNTER
  ARG_1_HI ARGUMENT_1_HI
  ARG_1_LO ARGUMENT_1_LO
  ARG_2_HI ARGUMENT_2_HI
  ARG_2_LO ARGUMENT_2_LO
  RES_HI   RESULT_HI
  RES_LO   RESULT_LO)

;; constants
(defconst
  ;; opcode values
  EVM_INST_SIGNEXTEND 0x0b
  EVM_INST_AND        0x16
  EVM_INST_OR         0x17
  EVM_INST_XOR        0x18
  EVM_INST_NOT        0x19
  EVM_INST_BYTE       0x1a
  ;; constant values
  LLARGE             16
  LLARGEMO           (- LLARGE 1))


(module bin)

;; 2.2  Shorthands
(defun (flag-sum)
  (+ IS_AND IS_OR IS_XOR IS_NOT IS_BYTE IS_SIGNEXTEND))

(defun (weight-sum)
  (+ (* IS_AND EVM_INST_AND)
     (* IS_OR EVM_INST_OR)
     (* IS_XOR EVM_INST_XOR)
     (* IS_NOT EVM_INST_NOT)
     (* IS_BYTE EVM_INST_BYTE)
     (* IS_SIGNEXTEND EVM_INST_SIGNEXTEND)))

;; 2.3 Instruction decoding
(defconstraint no-bin-no-flag ()
  (if (== STAMP 0)
      (== (flag-sum) 0)
      (== (flag-sum) 1)))

(defconstraint inst-to-flag ()
  (== INST (weight-sum)))

;; 2.4 Heartbeat
(defconstraint first-row (:domain {0})
  (== STAMP 0))

(defconstraint stamp-increments ()
  (∨ (will-inc! STAMP 0) (will-inc! STAMP 1)))

(defconstraint new-stamp-reset-ct ()
  (if (!= (next STAMP) STAMP)
      (== (next CT) 0)))

(defconstraint isnot-ctmax ()
  (if (== IS_NOT 1) (== CT_MAX LLARGEMO)))

(defconstraint isbyte-ctmax ()
  (if (== (+ IS_BYTE IS_SIGNEXTEND) 1)
         (if (== ARG_1_HI 0)
             (== CT_MAX LLARGEMO)
             (== CT_MAX 0))))

(defconstraint ct-small ()
  (!= CT LLARGE))

(defconstraint countereset (:guard STAMP)
  (if (== CT CT_MAX)
      (will-inc! STAMP 1)
      (will-inc! CT 1)))

(defconstraint last-row (:domain {-1})
  (== CT CT_MAX))

(defconstraint counter-constancies ()
  (begin (counter-constancy CT ARG_1_HI)
         (counter-constancy CT ARG_1_LO)
         (counter-constancy CT ARG_2_HI)
         (counter-constancy CT ARG_2_LO)
         (counter-constancy CT RES_HI)
         (counter-constancy CT RES_LO)
         (counter-constancy CT INST)
         (counter-constancy CT CT_MAX)
         (counter-constancy CT PIVOT)
         (counter-constancy CT BIT_B_4)
         (counter-constancy CT LOW_4)
         (counter-constancy CT NEG)
         (counter-constancy CT SMALL)))

;;    2.6 byte decompositions
(defconstraint byte_decompositions ()
  (begin (byte-decomposition CT ACC_1 BYTE_1)
         (byte-decomposition CT ACC_2 BYTE_2)
         (byte-decomposition CT ACC_3 BYTE_3)
         (byte-decomposition CT ACC_4 BYTE_4)
         (byte-decomposition CT ACC_5 BYTE_5)
         (byte-decomposition CT ACC_6 BYTE_6)))

;;    2.7 target constraints
(defun (requires-byte-decomposition)
  (+ IS_AND
     IS_OR
     IS_XOR
     IS_NOT
     (* CT_MAX (+ IS_BYTE IS_SIGNEXTEND))))

(defconstraint target-constraints (:guard (requires-byte-decomposition))
  (if (== CT CT_MAX)
      (begin (== ACC_1 ARG_1_HI)
             (== ACC_2 ARG_1_LO)
             (== ACC_3 ARG_2_HI)
             (== ACC_4 ARG_2_LO)
             (== ACC_5 RES_HI)
             (== ACC_6 RES_LO))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                                     ;;
;;    2.8 binary column constraints    ;;
;;                                     ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; 2.8.2 BITS and related columns
(defconstraint bits-and-related (:guard (+ IS_BYTE IS_SIGNEXTEND))
  (if (== CT LLARGEMO)
      (begin (== PIVOT
                 (+ (* 128 (shift BITS -15))
                    (* 64 (shift BITS -14))
                    (* 32 (shift BITS -13))
                    (* 16 (shift BITS -12))
                    (* 8 (shift BITS -11))
                    (* 4 (shift BITS -10))
                    (* 2 (shift BITS -9))
                    (shift BITS -8)))
             (== BYTE_2
                  (+ (* 128 (shift BITS -7))
                     (* 64 (shift BITS -6))
                     (* 32 (shift BITS -5))
                     (* 16 (shift BITS -4))
                     (* 8 (shift BITS -3))
                     (* 4 (shift BITS -2))
                     (* 2 (shift BITS -1))
                     BITS))
             (== LOW_4
                  (+ (* 8 (shift BITS -3))
                     (* 4 (shift BITS -2))
                     (* 2 (shift BITS -1))
                     BITS))
             (== BIT_B_4 (shift BITS -4))
             (== NEG (shift BITS -15)))))

;; 2.8.3 [[1]] constraints
(defconstraint bit_1 (:guard CT_MAX)
  (begin (if (== IS_BYTE 1)
             (plateau-constraint CT BIT_1 LOW_4))
         ;;
         (if (== IS_SIGNEXTEND 1)
             (plateau-constraint CT BIT_1 (- LLARGEMO LOW_4)))))

;; 2.8.4 SMALL constraints
(defconstraint small (:guard (+ IS_BYTE IS_SIGNEXTEND))
  (if (== CT LLARGEMO)
         (if (== ARG_1_LO (+ (* 16 (shift BITS -4))
                             (* 8 (shift BITS -3))
                             (* 4 (shift BITS -2))
                             (* 2 (shift BITS -1))
                             BITS))
             ;; if ARG_1_LO < 32, SMALL == 1
             (== SMALL 1)
             ;; if ARG_1_LO >= 32, SMALL == 0
             (== SMALL 0))))

;;    2.9 pivot constraints
(defconstraint pivot (:guard CT_MAX)
  (begin
   (if (== IS_BYTE 1)
       (if (== LOW_4 0)
           (if (== CT 0)
               (if (== BIT_B_4 0)
                   (== PIVOT BYTE_3)
                   (== PIVOT BYTE_4)))
           (if (== (+ (prev BIT_1) (- 1 BIT_1)) 0)
               (if (== BIT_B_4 0)
                   (== PIVOT BYTE_3)
                   (== PIVOT BYTE_4)))))
   (if (== IS_SIGNEXTEND 1)
       (if (== LOW_4 LLARGEMO)
           ;;
           (if (== CT 0)
               (if (== BIT_B_4 0)
                   (== PIVOT BYTE_4)
                   (== PIVOT BYTE_3)))
           ;; else
           (if (== (+ (prev BIT_1) (- 1 BIT_1)) 0)
               (if (== BIT_B_4 0)
                   (== PIVOT BYTE_4)
                   (== PIVOT BYTE_3)))))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                              ;;
;;    2.10 result constraints   ;;
;;                              ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
(defconstraint is-byte-result (:guard IS_BYTE)
  (if (== CT_MAX 0)
      (begin (== RES_HI 0)
             (== RES_LO 0))
      (begin (== RES_HI 0)
             (== RES_LO (* SMALL PIVOT)))))

(defconstraint is-signextend-result (:guard IS_SIGNEXTEND)
  (if (== CT_MAX 0)
      (begin (== RES_HI ARG_2_HI)
             (== RES_LO ARG_2_LO))
      (if (== SMALL 0)
          ;; SMALL == 0
          (begin (== RES_HI ARG_2_HI)
                 (== RES_LO ARG_2_LO))
          ;; SMALL == 1
          (begin (if (== BIT_B_4 0)
                     ;; b4 == 0
                     (begin (== BYTE_5 (* NEG 255))
                            (if (== BIT_1 0)
                                ;; [[1]] == 0
                                (== BYTE_6 (* NEG 255))
                                ;; [[1]] == 1
                                (== BYTE_6 BYTE_4)))
                     ;; b4 == 1
                     (begin
                      (if (== BIT_1 0)
                               ;; [[1]] == 0
                               (== BYTE_5 (* NEG 255))
                               ;; [[1]] == 1
                               (== BYTE_5 BYTE_3))
                      ;;
                      (== RES_LO ARG_2_LO)))))))

(defconstraint result-via-lookup (:guard (+ IS_AND IS_OR IS_XOR IS_NOT))
  (begin (== BYTE_5 XXX_BYTE_HI)
         (== BYTE_6 XXX_BYTE_LO)))
