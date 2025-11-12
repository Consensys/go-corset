(defconst
  EVM_INST_MUL                              0x02
  EVM_INST_EXP                              0x0a
  ;;
  MMEDIUM                                   8
  MMEDIUMMO                                 (- MMEDIUM 1))

(module mul)

(defcolumns
    (MUL_STAMP         :i32)
    (COUNTER           :byte)
    (OLI               :binary@prove)
    (TINY_BASE         :binary@prove)
    (TINY_EXPONENT     :binary@prove)
    (RESULT_VANISHES   :binary@prove)
    (INSTRUCTION       :byte :display :opcode)
    (ARG_1_HI          :i128)
    (ARG_1_LO          :i128)
    (ARG_2_HI          :i128)
    (ARG_2_LO          :i128)
    (RES_HI            :i128)
    (RES_LO            :i128)
    (BITS              :binary@prove)
    ;==========================
    (BYTE_A_3 :byte@prove)    (ACC_A_3 :i64)
    (BYTE_A_2 :byte@prove)    (ACC_A_2 :i64)
    (BYTE_A_1 :byte@prove)    (ACC_A_1 :i64)
    (BYTE_A_0 :byte@prove)    (ACC_A_0 :i64)
    ;==========================
    (BYTE_B_3 :byte@prove)    (ACC_B_3 :i64)
    (BYTE_B_2 :byte@prove)    (ACC_B_2 :i64)
    (BYTE_B_1 :byte@prove)    (ACC_B_1 :i64)
    (BYTE_B_0 :byte@prove)    (ACC_B_0 :i64)
    ;==========================
    (BYTE_C_3 :byte@prove)    (ACC_C_3 :i64)
    (BYTE_C_2 :byte@prove)    (ACC_C_2 :i64)
    (BYTE_C_1 :byte@prove)    (ACC_C_1 :i64)
    (BYTE_C_0 :byte@prove)    (ACC_C_0 :i64)
    ;==========================
    (BYTE_H_3 :byte@prove)    (ACC_H_3 :i64)
    (BYTE_H_2 :byte@prove)    (ACC_H_2 :i64)
    (BYTE_H_1 :byte@prove)    (ACC_H_1 :i64)
    (BYTE_H_0 :byte@prove)    (ACC_H_0 :i64)
    ;==========================
    (EXPONENT_BIT               :binary@prove)
    (EXPONENT_BIT_ACCUMULATOR   :i128)
    (EXPONENT_BIT_SOURCE        :binary@prove)
    (SQUARE_AND_MULTIPLY        :binary@prove)
    (BIT_NUM                    :i7)
)

(defalias

    STAMP       MUL_STAMP
    CT          COUNTER
    INST        INSTRUCTION
    EBIT        EXPONENT_BIT
    EACC        EXPONENT_BIT_ACCUMULATOR
    ESRC        EXPONENT_BIT_SOURCE
    SNM         SQUARE_AND_MULTIPLY
    TINYB       TINY_BASE
    TINYE       TINY_EXPONENT
    RESV        RESULT_VANISHES)
(module mul)

(defconst
  ONETWOEIGHT 128
  ONETWOSEVEN 127
  THETA       18446744073709551616                     ;18446744073709551616 = 256^8
  THETA2      340282366920938463463374607431768211456) ;340282366920938463463374607431768211456 = 256^16

;;;;;;;;;;;;;;;;;;;;;;;;;
;;                     ;;
;;    1.3 heartbeat    ;;
;;                     ;;
;;;;;;;;;;;;;;;;;;;;;;;;;

(defconstraint stamp-init (:domain {0}) ;; ""
  (vanishes! STAMP))

(defconstraint stamp-update ()
  (or! (will-inc! STAMP 1) (will-remain-constant! STAMP)))

(defconstraint vanishing ()
  (if-zero STAMP
           (begin (vanishes! CT)
                  (vanishes! OLI)
                  (vanishes! INST))))

(defconstraint stamp-constancies ()
  (begin (stamp-constancy STAMP ARG_1_HI)
         (stamp-constancy STAMP ARG_1_LO)
         (stamp-constancy STAMP ARG_2_HI)
         (stamp-constancy STAMP ARG_2_LO)
         (stamp-constancy STAMP RES_HI)
         (stamp-constancy STAMP RES_LO)
         (stamp-constancy STAMP INST)))

(defconstraint instruction-constraining ()
  (if-not-zero STAMP
               (vanishes! (* (- INST EVM_INST_MUL) (- INST EVM_INST_EXP)))))

(defconstraint reset-stuff ()
  (if-not (will-remain-constant! STAMP)
               (begin (vanishes! (next CT))
                      (vanishes! (next BIT_NUM)))))

(defconstraint oli-last-one-line ()
  (if-not-zero OLI
               (will-inc! STAMP 1)))

(defconstraint counter-update (:guard STAMP)
  (if-zero OLI
           (if-not-zero (- CT MMEDIUMMO)
                        (will-inc! CT 1))))

(defconstraint counter-reset ()
  (if-eq CT MMEDIUMMO
         (vanishes! (next CT))))

(defconstraint counter-constancies ()
  (begin (counter-constancy CT SNM)
         (counter-constancy CT BIT_NUM)
         (counter-constancy CT ESRC)
         (counter-constancy CT EBIT)
         (counter-constancy CT EACC)))

(defconstraint other-resets ()
  (if-eq CT MMEDIUMMO
         (begin (if-not-zero (- INST EVM_INST_EXP)
                             (will-inc! STAMP 1)) ; i.e. INST == MUL
                (if-not-zero (- INST EVM_INST_MUL)         ; i.e. INST == EXP
                             (if-eq RESV 1 (will-inc! STAMP 1))))))

(defconstraint bit_num-doesnt-reach-oneTwoEight ()
  (if-eq BIT_NUM ONETWOEIGHT (vanishes! 1)))

(defconstraint last-row (:domain {-1} :guard STAMP) ;; ""
  (begin (debug (eq! OLI 1))
         (eq! INST EVM_INST_EXP)
         (vanishes! ARG_1_HI)
         (vanishes! ARG_1_LO)
         (vanishes! ARG_2_HI)
         (vanishes! ARG_2_LO)
         (debug (eq! RES_HI 0))
         (debug (eq! RES_LO 1))))

(defun (first-row)
  (if-not-zero (- (prev STAMP) STAMP)
               (begin (eq! SNM 1)
                      (eq! EBIT 1)
                      (eq! EACC 1)
                      (if-zero ARG_2_HI
                               (eq! ESRC 1)
                               (vanishes! ESRC)))))

;; exponent-bit-source-is-high-limb applies when ESRC == 0
(defun (exponent-bit-source-is-high-limb)
  (begin (will-remain-constant! STAMP)
         (vanishes! (next SNM))
         (if-eq-else EACC ARG_2_HI
                     (begin (will-eq! ESRC 1)
                            (vanishes! (next EACC))
                            (vanishes! (next BIT_NUM)))
                     (begin (vanishes! (next ESRC))
                            (will-eq! EACC (* 2 EACC))
                            (will-inc! BIT_NUM 1)))))

;; exponent-bit-source-is-low-limb applies when ESRC == 1
(defun (exponent-bit-source-is-low-limb)
  (if-not-zero ARG_2_HI
               (if-eq-else BIT_NUM ONETWOSEVEN
                           ;; (ARG_2_HI != 0) et (BIT_NUM == 127)
                           (begin (will-inc! STAMP 1)
                                  (eq! EACC ARG_2_LO))
                           ;; (ARG_2_HI != 0) et (BIT_NUM != 127)
                           (begin (vanishes! (next SNM))
                                  (will-remain-constant! STAMP)
                                  (will-eq! ESRC 1)
                                  (will-eq! EACC (* 2 EACC))
                                  (will-inc! BIT_NUM 1)))
               (if-eq-else EACC ARG_2_LO
                           ;; (ARG_2_HI == 0) et (EACC == ARG_2_LO)
                           (will-inc! STAMP 1)
                           ;; (ARG_2_HI == 0) et (EACC != ARG_2_LO)
                           (begin (vanishes! (next SNM))
                                  (will-remain-constant! STAMP)
                                  (will-eq! ESRC 1)
                                  (will-eq! EACC (* 2 EACC))
                                  (will-inc! BIT_NUM 1)))))

(defun (end-of-cycle)
  (if-zero (- SNM EBIT)
           ;; SNM == EBIT
           (if-zero ESRC
                    ;; ESRC == 0 i.e. source = high part
                    (exponent-bit-source-is-high-limb)
                    ;; ESRC == 1 i.e. source = low part
                    (exponent-bit-source-is-low-limb))
           ;; SNM != EBIT
           (begin (will-remain-constant! STAMP)
                  (will-inc! SNM 1)
                  (will-remain-constant! EBIT)
                  (will-inc! EACC 1)
                  (will-remain-constant! ESRC)
                  (will-remain-constant! BIT_NUM))))

(defconstraint nontrivial-exp-regime-nonzero-result-heartbeat ()
  (if-eq INST EVM_INST_EXP
         (if-zero OLI
                  (if-zero RESV
                           (begin (first-row)
                                  (if-eq CT MMEDIUMMO (end-of-cycle)))))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                               ;;
;;    1.5 byte decompositions    ;;
;;                               ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defconstraint byte-decompositions ()
  (begin (byte-decomposition CT ACC_A_0 BYTE_A_0)
         (byte-decomposition CT ACC_A_1 BYTE_A_1)
         (byte-decomposition CT ACC_A_2 BYTE_A_2)
         (byte-decomposition CT ACC_A_3 BYTE_A_3)
         ;
         (byte-decomposition CT ACC_B_0 BYTE_B_0)
         (byte-decomposition CT ACC_B_1 BYTE_B_1)
         (byte-decomposition CT ACC_B_2 BYTE_B_2)
         (byte-decomposition CT ACC_B_3 BYTE_B_3)
         ;
         (byte-decomposition CT ACC_C_0 BYTE_C_0)
         (byte-decomposition CT ACC_C_1 BYTE_C_1)
         (byte-decomposition CT ACC_C_2 BYTE_C_2)
         (byte-decomposition CT ACC_C_3 BYTE_C_3)
         ;
         (byte-decomposition CT ACC_H_0 BYTE_H_0)
         (byte-decomposition CT ACC_H_1 BYTE_H_1)
         (byte-decomposition CT ACC_H_2 BYTE_H_2)
         (byte-decomposition CT ACC_H_3 BYTE_H_3)))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                                                  ;;
;;    1.6 TINYB, TINYE, OLI and RESV constraints    ;;
;;                                                  ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defconstraint tiny-base (:guard STAMP)
  (if-not-zero ARG_1_HI
               (vanishes! TINYB)
               (if-not-zero (* ARG_1_LO (- 1 ARG_1_LO))
                            (vanishes! TINYB)
                            (eq! TINYB 1))))

(defconstraint tiny-exponent (:guard STAMP)
  (if-not-zero ARG_2_HI
               (vanishes! TINYE)
               (if-not-zero (* ARG_2_LO (- 1 ARG_2_LO))
                            (vanishes! TINYE)
                            (eq! TINYE 1))))

(defconstraint result-vanishes! (:guard STAMP)
  (if-not-zero RES_HI
               (vanishes! RESV)
               (if-not-zero RES_LO
                            (vanishes! RESV)
                            (eq! RESV 1))))

(defconstraint one-line-instruction (:guard STAMP)
  (eq! (+ OLI (* TINYB TINYE))
     (+ TINYB TINYE)))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                          ;;
;;    1.7 trivial regime    ;;
;;                          ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defconstraint trivial-regime ()
  (if (== OLI 1)
         ;; since OLI != 0 we have STAMP != 0
         ;; thus INST ∈ {MUL, EXP}
         (begin (if (!= INST EVM_INST_EXP)
                             ;; i.e. INST == MUL
                             (begin (if (== TINYE 1)
                                           ;; i.e. ARG_2 = ARG_2_LO ∈ {0, 1}
                                           (begin (eq! RES_HI (* ARG_2_LO ARG_1_HI))
                                                  (eq! RES_LO (* ARG_2_LO ARG_1_LO))))
                                    (if (== TINYB 1)
                                           ;; i.e. ARG_1 = ARG_1_LO ∈ {0, 1}
                                           (begin (eq! RES_HI (* ARG_1_LO ARG_2_HI))
                                                  (eq! RES_LO (* ARG_1_LO ARG_2_LO))))))
                (if (!= INST EVM_INST_MUL)
                    ;; i.e. INST == EXP
                    (begin (if (== TINYE 1)
                               ;; TINYE == 1 <=> ARG_2 = ARG_2_LO ∈ {0, 1}
                               (begin (if (!= ARG_2_LO 1)
                                          ;; Thus ARG_2_LO != 1 <=> ARG_2_LO == 0
                                          (begin (== 0 RES_HI)
                                                 (== RES_LO 1)))
                                      (if (!= ARG_2_LO 0)
                                          ;; Thus ARG_2_LO != 0 <=> ARG_2_LO == 1
                                          (begin (== RES_HI ARG_1_HI)
                                                 (== RES_LO ARG_1_LO))))
                               ;; TINYE == 0 but OLI == 1 thus TINYB == 1
                               (begin (== RES_HI ARG_1_HI)
                                      (== RES_LO ARG_1_LO))))))))

;;;;;;;;;;;;;;;;;;;;;;;
;;                   ;;
;;    1.8 aliases    ;;
;;                   ;;
;;;;;;;;;;;;;;;;;;;;;;;

(defun (A_3) ACC_A_3)
(defun (A_2) ACC_A_2)
(defun (A_1) ACC_A_1)
(defun (A_0) ACC_A_0)

;========
(defun (B_3) ACC_B_3)
(defun (B_2) ACC_B_2)
(defun (B_1) ACC_B_1)
(defun (B_0) ACC_B_0)

;========
(defun (C_3) ACC_C_3)
(defun (C_2) ACC_C_2)
(defun (C_1) ACC_C_1)
(defun (C_0) ACC_C_0)

;========
(defun (C'_3) (shift ACC_C_3 -8))
(defun (C'_2) (shift ACC_C_2 -8))
(defun (C'_1) (shift ACC_C_1 -8))
(defun (C'_0) (shift ACC_C_0 -8))

;========
(defun (H_3) ACC_H_3)
(defun (H_2) ACC_H_2)
(defun (H_1) ACC_H_1)
(defun (H_0) ACC_H_0)

;========
(defun (alpha)    (shift BITS -5))
(defun (beta_0)   (shift BITS -4))
(defun (beta_1)   (shift BITS -3))
(defun (eta)      (shift BITS -2))
(defun (mu_0)     (shift BITS -1))
(defun (mu_1)            BITS)

;========
(defun (beta) (+ (* 2 (beta_1)) (beta_0)))
(defun (mu)   (+ (* 2 (mu_1))   (mu_0)))    ;; ""

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                                 ;;
;;    1.9 nontrivial MUL regime    ;;
;;                                 ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defconstraint nontrivial-mul-regime ()
  (if-eq CT MMEDIUMMO
         ;; i.e. if INST == MUL
         (if-not-zero (- INST EVM_INST_EXP)
                      ;; byte decomposition constraints
                      (begin (eq! ARG_1_HI
                                (+ (* THETA (A_3)) (A_2)))
                             (eq! ARG_1_LO
                                (+ (* THETA (A_1)) (A_0)))
                             (eq! ARG_2_HI
                                (+ (* THETA (B_3)) (B_2)))
                             (eq! ARG_2_LO
                                (+ (* THETA (B_1)) (B_0)))
                             (eq! RES_HI
                                (+ (* THETA (C_3)) (C_2)))
                             (eq! RES_LO
                                (+ (* THETA (C_1)) (C_0)))
                             ;; multiplication per se
                             (set-multiplication (A_3)
                                                 (A_2)
                                                 (A_1)
                                                 (A_0)
                                                 (B_3)
                                                 (B_2)
                                                 (B_1)
                                                 (B_0)
                                                 (C_3)
                                                 (C_2)
                                                 (C_1)
                                                 (C_0)
                                                 (H_3)
                                                 (H_2)
                                                 (H_1)
                                                 (H_0)
                                                 (alpha)
                                                 (beta)
                                                 (eta)
                                                 (mu))))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                                                ;;
;;    1.10 nontrivial EXP regime - zero result    ;;
;;                                                ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defun (special-constraints-for-byte-c-0)
  (if-eq-else CT MMEDIUMMO
              (if-zero (A_0)
                       (vanishes! BYTE_C_0)
                       (eq! BYTE_C_0 1))
              (will-remain-constant! BYTE_C_0)))

(defun (preparations-for-a-lower-bound-on-the-2-adicity-of-the-base)
  (begin  ;; recall that BYTE_C_0 will be either 0 or 1 !
         ;; cf. (special-constraints-for-byte-c-0)
         (if-not-zero BYTE_C_0
                      (prepare-lower-bound-on-two-adicity BYTE_A_0
                                                          BYTE_C_1
                                                          BITS
                                                          BYTE_C_3
                                                          BYTE_H_3
                                                          BYTE_C_2
                                                          BYTE_H_2
                                                          CT))
         (if-not-zero (- 1 BYTE_C_0)
                      (prepare-lower-bound-on-two-adicity BYTE_A_1
                                                          BYTE_C_1
                                                          BITS
                                                          BYTE_C_3
                                                          BYTE_H_3
                                                          BYTE_C_2
                                                          BYTE_H_2
                                                          CT))))

(defun (nu2-byte-c-0-equals-1) (+ (* 8 BYTE_H_3) BYTE_H_2))
(defun (nu2-byte-c-0-equals-0) (+ (* 8 BYTE_H_3) BYTE_H_2 (* 8 MMEDIUM)))
;; θ · (B_3 + B_2 + B_1) + B_0
(defun (test-for-bytehood-of-arg-2) (+ (B_0)
                                       (* THETA (+ (B_3) (B_2) (B_1)))))

(defun (proving-the-vanishing-of-exp)
  (if-eq CT MMEDIUMMO
         (if-eq-else (test-for-bytehood-of-arg-2) BYTE_B_0
                     ;; ARG_2 is a byte
                     (begin (if-not-zero BYTE_C_0
                                         (eq! (* (B_0) (nu2-byte-c-0-equals-1)) (+ 256 (H_1))))
                            (if-not-zero (- 1 BYTE_C_0)
                                         (eq! (* (B_0) (nu2-byte-c-0-equals-0)) (+ 256 (H_1)))))
                     ;; ARG_2 isn't a byte
                     (begin (if-not-zero BYTE_C_0
                                         (eq! (nu2-byte-c-0-equals-1) (+ 1 (H_1))))
                            (debug (if-not-zero (- 1 BYTE_C_0)
                                                (eq! (nu2-byte-c-0-equals-0) (+ 1 (H_1)))))))))

(defconstraint prepare-lower-bound-on-the-2-adicity-of-the-base ()
  ;; sincde we will later impose RESV == 1 we will have
  ;; STAMP != 0 and thus INST ∈ {MUL, EXP}
  ;; INST != MUL <=> INST == EXP
  (if-not-zero (- INST EVM_INST_MUL)
               (if-zero OLI
                        (if-not-zero RESV
                                     ;; target constraints
                                     (begin (if-eq CT MMEDIUMMO
                                                   (begin (eq! ARG_1_HI
                                                             (+ (* THETA (A_3)) (A_2)))
                                                          (eq! ARG_1_LO
                                                             (+ (* THETA (A_1)) (A_0)))
                                                          ;;
                                                          (eq! ARG_2_HI
                                                             (+ (* THETA (B_3)) (B_2)))
                                                          (eq! ARG_2_LO
                                                             (+ (* THETA (B_1)) (B_0)))))
                                            (if-not-zero ARG_1_LO
                                                         (begin (special-constraints-for-byte-c-0)
                                                                (preparations-for-a-lower-bound-on-the-2-adicity-of-the-base)
                                                                (proving-the-vanishing-of-exp))))))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;                                                   ;;
;;    1.11 nontrivial EXP regime - nonzero result    ;;
;;                                                   ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defun (target-arg1)
  (begin (eq! ARG_1_HI
            (+ (* THETA (A_3)) (A_2)))
         (eq! ARG_1_LO
            (+ (* THETA (A_1)) (A_0)))))

(defun (first-square-and-multiply)
  (if-not-zero (- (shift STAMP -8) STAMP)
               (begin (eq! ARG_1_HI
                         (+ (* THETA (C_3)) (C_2)))
                      (eq! ARG_1_LO
                         (+ (* THETA (C_1)) (C_0))))))

(defun (subsequent-square-and-multiply)
  (if-eq (shift STAMP -8) STAMP
         (if-zero SNM
                  ;; SQUARING
                  (set-multiplication (C'_3)
                                      (C'_2)
                                      (C'_1)
                                      (C'_0)
                                      (C'_3)
                                      (C'_2)
                                      (C'_1)
                                      (C'_0)
                                      (C_3)
                                      (C_2)
                                      (C_1)
                                      (C_0)
                                      (H_3)
                                      (H_2)
                                      (H_1)
                                      (H_0)
                                      (alpha)
                                      (beta)
                                      (eta)
                                      (mu))
                  ;; MULTIPLY
                  (set-multiplication (C'_3)
                                      (C'_2)
                                      (C'_1)
                                      (C'_0)
                                      (A_3)
                                      (A_2)
                                      (A_1)
                                      (A_0)
                                      (C_3)
                                      (C_2)
                                      (C_1)
                                      (C_0)
                                      (H_3)
                                      (H_2)
                                      (H_1)
                                      (H_0)
                                      (alpha)
                                      (beta)
                                      (eta)
                                      (mu)))))

(defun (final-square-and-multiply)
  (if-not (will-remain-constant! STAMP)
               (begin (eq! RES_HI
                         (+ (* THETA (C_3)) (C_2)))
                      (eq! RES_LO
                         (+ (* THETA (C_1)) (C_0))))))

(defconstraint nontrivial-exp-regime-nonzero-result ()
  (if-eq INST EVM_INST_EXP
         (if-zero RESV
                  (if-eq CT MMEDIUMMO
                         (begin (target-arg1)
                                (first-square-and-multiply)
                                (subsequent-square-and-multiply)
                                (final-square-and-multiply))))))
(module mul)

(defpurefun (set-multiplication
                a_3 a_2 a_1 a_0
                b_3 b_2 b_1 b_0
                p_3 p_2 p_1 p_0
                h_3 h_2 h_1 h_0
                alpha beta eta mu)
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
                (begin
                    ;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
                    (eq!  (+ (* a_1 b_0) (* a_0 b_1))
                        (+ (* THETA2 alpha) (* THETA h_1) h_0))
                    ;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
                    (eq!  (+ (* a_3 b_0) (* a_2 b_1) (* a_1 b_2) (* a_0 b_3))
                        (+ (* THETA2 beta) (* THETA h_3) h_2))
                    ;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
                    (eq!  (+ (* a_0 b_0) (* THETA h_0))
                        (+ (* THETA2 eta) (* THETA p_1) p_0))
                    ;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
                    (eq!  (+ eta h_1 (* THETA alpha) (* a_2 b_0) (* a_1 b_1) (* a_0 b_2) (* THETA h_2))
                        (+ (* THETA2 mu) (* THETA p_3) p_2))))


(defpurefun (prepare-lower-bound-on-two-adicity
                bytes cst bits
                x sumx y sumy ct)
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
                    (begin
                        (running-total x sumx ct)
                        (running-total y sumy ct)
                        (if-not-zero x (vanishes! bytes))    ; (see REMARK)
                        (if-not-zero y (vanishes! bits))     ; (see REMARK)
                        (if-not-zero (- ct MMEDIUMMO) (will-remain-constant! cst))
                        (if-not-zero (- ct MMEDIUMMO)
                            (if-not-zero (- 1 x)            ; (see REMARK)
                                (if-not-zero (next x)       ; (see REMARK)
                                    (eq! cst bytes))))
                        (if-eq ct MMEDIUMMO
                            (begin
                                (if-not-zero (- 1 x)        ; (see REMARK)
                                    (eq! cst bytes))
                                (eq! cst (bit-decomposition-of-byte bits))))))
;; REMARK:
;; within the scope of prepare-lower-bound-on-two-adicity
;; the running-total applies so that x and y are forced to
;; be binary (if only for the counter-cycle where the above applies)
;; in any case: 1 - x != 0 <=> x = 0

(defpurefun (bit-decomposition-of-byte bits)
                (+  (* 128  (shift bits -7))
                    (* 64   (shift bits -6))
                    (* 32   (shift bits -5))
                    (* 16   (shift bits -4))
                    (* 8    (shift bits -3))
                    (* 4    (shift bits -2))
                    (* 2    (shift bits -1))
                    bits))

(defpurefun (running-total x sumx ct)
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
                    (begin
                        (is-binary x)
                        (if-zero ct
                            (begin
                                (vanishes! x)
                                (vanishes! sumx)))
                        (if-not-zero (- ct MMEDIUMMO)
                            (begin
                             (or! (will-remain-constant! x) (will-inc! x 1))
                             (will-eq! sumx (+ sumx (next x)))))))
