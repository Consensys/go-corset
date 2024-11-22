(module wcp)

(defcolumns
  (WORD_COMPARISON_STAMP :i32)
  (COUNTER :byte)
  (CT_MAX :byte)
  (INST :byte)
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

(defconstraint result () (begin (if (- ONE_LINE_INSTRUCTION 1) (- RESULT (* BIT_1 BIT_2))) (if (- IS_LT 1) (- RESULT (- 1 (* BIT_1 BIT_2) (+ BIT_3 (* BIT_1 BIT_4))))) (if (- IS_GT 1) (- RESULT (+ BIT_3 (* BIT_1 BIT_4)))) (if (- IS_LEQ 1) (- RESULT (+ (- 1 (* BIT_1 BIT_2) (+ BIT_3 (* BIT_1 BIT_4))) (* BIT_1 BIT_2)))) (if (- IS_GEQ 1) (- RESULT (+ (+ BIT_3 (* BIT_1 BIT_4)) (* BIT_1 BIT_2)))) (if (- IS_LT 1) (- RESULT (- 1 (* BIT_1 BIT_2) (+ BIT_3 (* BIT_1 BIT_4))))) (if (- IS_SLT 1) (if (- NEG_1 NEG_2) (- RESULT (- 1 (* BIT_1 BIT_2) (+ BIT_3 (* BIT_1 BIT_4)))) (- RESULT NEG_1))) (if (- IS_SGT 1) (if (- NEG_1 NEG_2) (- RESULT (+ BIT_3 (* BIT_1 BIT_4))) (- RESULT NEG_2)))))

(defconstraint bits-and-negs () (if (+ IS_SLT IS_SGT) 0 (if (- COUNTER 15) (begin (- (shift BYTE_1 -15) (+ (* 1 (shift BITS -8)) (+ (* 2 (shift BITS -9)) (+ (* 4 (shift BITS -10)) (+ (* 8 (shift BITS -11)) (+ (* 16 (shift BITS -12)) (+ (* 32 (shift BITS -13)) (+ (* 128 (shift BITS -15)) (* 64 (shift BITS -14)))))))))) (- (shift BYTE_3 -15) (+ (* 1 BITS) (+ (* 2 (shift BITS -1)) (+ (* 4 (shift BITS -2)) (+ (* 8 (shift BITS -3)) (+ (* 16 (shift BITS -4)) (+ (* 32 (shift BITS -5)) (+ (* 128 (shift BITS -7)) (* 64 (shift BITS -6)))))))))) (- NEG_1 (shift BITS -15)) (- NEG_2 (shift BITS -7))))))

(defconstraint byte_decompositions () (begin (if COUNTER (- ACC_1 BYTE_1) (- ACC_1 (+ (* 256 (shift ACC_1 -1)) BYTE_1))) (if COUNTER (- ACC_2 BYTE_2) (- ACC_2 (+ (* 256 (shift ACC_2 -1)) BYTE_2))) (if COUNTER (- ACC_3 BYTE_3) (- ACC_3 (+ (* 256 (shift ACC_3 -1)) BYTE_3))) (if COUNTER (- ACC_4 BYTE_4) (- ACC_4 (+ (* 256 (shift ACC_4 -1)) BYTE_4))) (if COUNTER (- ACC_5 BYTE_5) (- ACC_5 (+ (* 256 (shift ACC_5 -1)) BYTE_5))) (if COUNTER (- ACC_6 BYTE_6) (- ACC_6 (+ (* 256 (shift ACC_6 -1)) BYTE_6)))))

(defconstraint target-constraints () (begin (if WORD_COMPARISON_STAMP 0 (begin (if (- ARGUMENT_1_HI ARGUMENT_2_HI) (- BIT_1 1) BIT_1) (if (- ARGUMENT_1_LO ARGUMENT_2_LO) (- BIT_2 1) BIT_2))) (if (- VARIABLE_LENGTH_INSTRUCTION 1) (if (- COUNTER CT_MAX) (begin (- ACC_1 ARGUMENT_1_HI) (- ACC_2 ARGUMENT_1_LO) (- ACC_3 ARGUMENT_2_HI) (- ACC_4 ARGUMENT_2_LO) (- ACC_5 (- (* (- (* 2 BIT_3) 1) (- ARGUMENT_1_HI ARGUMENT_2_HI)) BIT_3)) (- ACC_6 (- (* (- (* 2 BIT_4) 1) (- ARGUMENT_1_LO ARGUMENT_2_LO)) BIT_4))))) (if (- IS_ISZERO 1) (begin ARGUMENT_2_HI ARGUMENT_2_LO))))

(defconstraint counter-constancies () (begin (if COUNTER 0 (- ARGUMENT_1_HI (shift ARGUMENT_1_HI -1))) (if COUNTER 0 (- ARGUMENT_1_LO (shift ARGUMENT_1_LO -1))) (if COUNTER 0 (- ARGUMENT_2_HI (shift ARGUMENT_2_HI -1))) (if COUNTER 0 (- ARGUMENT_2_LO (shift ARGUMENT_2_LO -1))) (if COUNTER 0 (- RESULT (shift RESULT -1))) (if COUNTER 0 (- INST (shift INST -1))) (if COUNTER 0 (- CT_MAX (shift CT_MAX -1))) (if COUNTER 0 (- BIT_3 (shift BIT_3 -1))) (if COUNTER 0 (- BIT_4 (shift BIT_4 -1))) (if COUNTER 0 (- NEG_1 (shift NEG_1 -1))) (if COUNTER 0 (- NEG_2 (shift NEG_2 -1)))))

(defconstraint setting-flag () (begin (- INST (+ (* 16 IS_LT) (* 17 IS_GT) (* 18 IS_SLT) (* 19 IS_SGT) (* 20 IS_EQ) (* 21 IS_ISZERO) (* 15 IS_GEQ) (* 14 IS_LEQ))) (- ONE_LINE_INSTRUCTION (+ IS_EQ IS_ISZERO)) (- VARIABLE_LENGTH_INSTRUCTION (+ IS_LT IS_GT IS_LEQ IS_GEQ IS_SLT IS_SGT))))

(defconstraint inst-decoding () (if WORD_COMPARISON_STAMP (+ (+ IS_EQ IS_ISZERO) (+ IS_LT IS_GT IS_LEQ IS_GEQ IS_SLT IS_SGT)) (- (+ (+ IS_EQ IS_ISZERO) (+ IS_LT IS_GT IS_LEQ IS_GEQ IS_SLT IS_SGT)) 1)))

(defconstraint heartbeat () (if WORD_COMPARISON_STAMP 0 (if (- COUNTER CT_MAX) (- (shift WORD_COMPARISON_STAMP 1) (+ WORD_COMPARISON_STAMP 1)) (- (shift COUNTER 1) (+ COUNTER 1)))))

(defconstraint stamp-increments () (* (- (shift WORD_COMPARISON_STAMP 1) WORD_COMPARISON_STAMP) (- (shift WORD_COMPARISON_STAMP 1) (+ WORD_COMPARISON_STAMP 1))))

(defconstraint ct-upper-bond () (- (~ (- 16 COUNTER)) 1))

(defconstraint counter-reset () (if (- (shift WORD_COMPARISON_STAMP 1) WORD_COMPARISON_STAMP) 0 (shift COUNTER 1)))

(defconstraint setting-ct-max () (if (- ONE_LINE_INSTRUCTION 1) CT_MAX))

(defconstraint no-neg-if-small () (if (- CT_MAX 15) 0 (begin NEG_1 NEG_2)))

(defconstraint lastRow (:domain {-1}) (- COUNTER CT_MAX))

(defconstraint first-row (:domain {0}) WORD_COMPARISON_STAMP)
