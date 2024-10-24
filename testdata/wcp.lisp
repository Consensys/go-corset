(defcolumns
  (wcp:INST :u8)
  (wcp:ACC_1 :u128)
  (wcp:ARGUMENT_2_HI :u128)
  (wcp:ACC_3 :u128)
  (wcp:ARGUMENT_1_HI :u128)
  (wcp:ACC_6 :u128)
  (wcp:IS_GT :u1)
  (wcp:IS_GEQ :u1)
  (wcp:BIT_4 :u1)
  (wcp:BYTE_2 :u8)
  (wcp:IS_ISZERO :u1)
  (wcp:IS_LT :u1)
  (wcp:BITS :u1)
  (wcp:BIT_3 :u1)
  (wcp:ONE_LINE_INSTRUCTION :u1)
  (wcp:BYTE_5 :u8)
  (wcp:IS_EQ :u1)
  (wcp:BIT_1 :u1)
  (wcp:ARGUMENT_1_LO :u128)
  (wcp:BIT_2 :u1)
  (wcp:BYTE_1 :u8)
  (wcp:NEG_2 :u1)
  (wcp:BYTE_6 :u8)
  (wcp:CT_MAX :u8)
  (wcp:IS_LEQ :u1)
  (wcp:IS_SLT :u1)
  (wcp:ACC_4 :u128)
  (wcp:ARGUMENT_2_LO :u128)
  (wcp:ACC_2 :u128)
  (wcp:WORD_COMPARISON_STAMP :u32)
  (wcp:RESULT :u1)
  (wcp:COUNTER :u8)
  (wcp:NEG_1 :u1)
  (wcp:ACC_5 :u128)
  (wcp:IS_SGT :u1)
  (wcp:BYTE_4 :u8)
  (wcp:BYTE_3 :u8)
  (wcp:VARIABLE_LENGTH_INSTRUCTION :u1))

(defconstraint wcp:result () (begin (if (- wcp:ONE_LINE_INSTRUCTION 1) (- wcp:RESULT (* wcp:BIT_1 wcp:BIT_2))) (if (- wcp:IS_LT 1) (- wcp:RESULT (- 1 (* wcp:BIT_1 wcp:BIT_2) (+ wcp:BIT_3 (* wcp:BIT_1 wcp:BIT_4))))) (if (- wcp:IS_GT 1) (- wcp:RESULT (+ wcp:BIT_3 (* wcp:BIT_1 wcp:BIT_4)))) (if (- wcp:IS_LEQ 1) (- wcp:RESULT (+ (- 1 (* wcp:BIT_1 wcp:BIT_2) (+ wcp:BIT_3 (* wcp:BIT_1 wcp:BIT_4))) (* wcp:BIT_1 wcp:BIT_2)))) (if (- wcp:IS_GEQ 1) (- wcp:RESULT (+ (+ wcp:BIT_3 (* wcp:BIT_1 wcp:BIT_4)) (* wcp:BIT_1 wcp:BIT_2)))) (if (- wcp:IS_LT 1) (- wcp:RESULT (- 1 (* wcp:BIT_1 wcp:BIT_2) (+ wcp:BIT_3 (* wcp:BIT_1 wcp:BIT_4))))) (if (- wcp:IS_SLT 1) (if (- wcp:NEG_1 wcp:NEG_2) (- wcp:RESULT (- 1 (* wcp:BIT_1 wcp:BIT_2) (+ wcp:BIT_3 (* wcp:BIT_1 wcp:BIT_4)))) (- wcp:RESULT wcp:NEG_1))) (if (- wcp:IS_SGT 1) (if (- wcp:NEG_1 wcp:NEG_2) (- wcp:RESULT (+ wcp:BIT_3 (* wcp:BIT_1 wcp:BIT_4))) (- wcp:RESULT wcp:NEG_2)))))

(defconstraint wcp:bits-and-negs () (ifnot (+ wcp:IS_SLT wcp:IS_SGT) (if (- wcp:COUNTER 15) (begin (- (shift wcp:BYTE_1 -15) (+ (* 1 (shift wcp:BITS -8)) (+ (* 2 (shift wcp:BITS -9)) (+ (* 4 (shift wcp:BITS -10)) (+ (* 8 (shift wcp:BITS -11)) (+ (* 16 (shift wcp:BITS -12)) (+ (* 32 (shift wcp:BITS -13)) (+ (* 128 (shift wcp:BITS -15)) (* 64 (shift wcp:BITS -14)))))))))) (- (shift wcp:BYTE_3 -15) (+ (* 1 wcp:BITS) (+ (* 2 (shift wcp:BITS -1)) (+ (* 4 (shift wcp:BITS -2)) (+ (* 8 (shift wcp:BITS -3)) (+ (* 16 (shift wcp:BITS -4)) (+ (* 32 (shift wcp:BITS -5)) (+ (* 128 (shift wcp:BITS -7)) (* 64 (shift wcp:BITS -6)))))))))) (- wcp:NEG_1 (shift wcp:BITS -15)) (- wcp:NEG_2 (shift wcp:BITS -7))))))

(defconstraint wcp:byte_decompositions () (begin (if wcp:COUNTER (- wcp:ACC_1 wcp:BYTE_1) (- wcp:ACC_1 (+ (* 256 (shift wcp:ACC_1 -1)) wcp:BYTE_1))) (if wcp:COUNTER (- wcp:ACC_2 wcp:BYTE_2) (- wcp:ACC_2 (+ (* 256 (shift wcp:ACC_2 -1)) wcp:BYTE_2))) (if wcp:COUNTER (- wcp:ACC_3 wcp:BYTE_3) (- wcp:ACC_3 (+ (* 256 (shift wcp:ACC_3 -1)) wcp:BYTE_3))) (if wcp:COUNTER (- wcp:ACC_4 wcp:BYTE_4) (- wcp:ACC_4 (+ (* 256 (shift wcp:ACC_4 -1)) wcp:BYTE_4))) (if wcp:COUNTER (- wcp:ACC_5 wcp:BYTE_5) (- wcp:ACC_5 (+ (* 256 (shift wcp:ACC_5 -1)) wcp:BYTE_5))) (if wcp:COUNTER (- wcp:ACC_6 wcp:BYTE_6) (- wcp:ACC_6 (+ (* 256 (shift wcp:ACC_6 -1)) wcp:BYTE_6)))))

(defconstraint wcp:target-constraints () (begin (ifnot wcp:WORD_COMPARISON_STAMP (begin (if (- wcp:ARGUMENT_1_HI wcp:ARGUMENT_2_HI) (- wcp:BIT_1 1) wcp:BIT_1) (if (- wcp:ARGUMENT_1_LO wcp:ARGUMENT_2_LO) (- wcp:BIT_2 1) wcp:BIT_2))) (if (- wcp:VARIABLE_LENGTH_INSTRUCTION 1) (if (- wcp:COUNTER wcp:CT_MAX) (begin (- wcp:ACC_1 wcp:ARGUMENT_1_HI) (- wcp:ACC_2 wcp:ARGUMENT_1_LO) (- wcp:ACC_3 wcp:ARGUMENT_2_HI) (- wcp:ACC_4 wcp:ARGUMENT_2_LO) (- wcp:ACC_5 (- (* (- (* 2 wcp:BIT_3) 1) (- wcp:ARGUMENT_1_HI wcp:ARGUMENT_2_HI)) wcp:BIT_3)) (- wcp:ACC_6 (- (* (- (* 2 wcp:BIT_4) 1) (- wcp:ARGUMENT_1_LO wcp:ARGUMENT_2_LO)) wcp:BIT_4))))) (if (- wcp:IS_ISZERO 1) (begin wcp:ARGUMENT_2_HI wcp:ARGUMENT_2_LO))))

(defconstraint wcp:counter-constancies () (begin (ifnot wcp:COUNTER (- wcp:ARGUMENT_1_HI (shift wcp:ARGUMENT_1_HI -1))) (ifnot wcp:COUNTER (- wcp:ARGUMENT_1_LO (shift wcp:ARGUMENT_1_LO -1))) (ifnot wcp:COUNTER (- wcp:ARGUMENT_2_HI (shift wcp:ARGUMENT_2_HI -1))) (ifnot wcp:COUNTER (- wcp:ARGUMENT_2_LO (shift wcp:ARGUMENT_2_LO -1))) (ifnot wcp:COUNTER (- wcp:RESULT (shift wcp:RESULT -1))) (ifnot wcp:COUNTER (- wcp:INST (shift wcp:INST -1))) (ifnot wcp:COUNTER (- wcp:CT_MAX (shift wcp:CT_MAX -1))) (ifnot wcp:COUNTER (- wcp:BIT_3 (shift wcp:BIT_3 -1))) (ifnot wcp:COUNTER (- wcp:BIT_4 (shift wcp:BIT_4 -1))) (ifnot wcp:COUNTER (- wcp:NEG_1 (shift wcp:NEG_1 -1))) (ifnot wcp:COUNTER (- wcp:NEG_2 (shift wcp:NEG_2 -1)))))

(defconstraint wcp:setting-flag () (begin (- wcp:INST (+ (* 16 wcp:IS_LT) (* 17 wcp:IS_GT) (* 18 wcp:IS_SLT) (* 19 wcp:IS_SGT) (* 20 wcp:IS_EQ) (* 21 wcp:IS_ISZERO) (* 15 wcp:IS_GEQ) (* 14 wcp:IS_LEQ))) (- wcp:ONE_LINE_INSTRUCTION (+ wcp:IS_EQ wcp:IS_ISZERO)) (- wcp:VARIABLE_LENGTH_INSTRUCTION (+ wcp:IS_LT wcp:IS_GT wcp:IS_LEQ wcp:IS_GEQ wcp:IS_SLT wcp:IS_SGT))))

(defconstraint wcp:inst-decoding () (if wcp:WORD_COMPARISON_STAMP (+ (+ wcp:IS_EQ wcp:IS_ISZERO) (+ wcp:IS_LT wcp:IS_GT wcp:IS_LEQ wcp:IS_GEQ wcp:IS_SLT wcp:IS_SGT)) (- (+ (+ wcp:IS_EQ wcp:IS_ISZERO) (+ wcp:IS_LT wcp:IS_GT wcp:IS_LEQ wcp:IS_GEQ wcp:IS_SLT wcp:IS_SGT)) 1)))

(defconstraint wcp:heartbeat () (ifnot wcp:WORD_COMPARISON_STAMP (if (- wcp:COUNTER wcp:CT_MAX) (- (shift wcp:WORD_COMPARISON_STAMP 1) (+ wcp:WORD_COMPARISON_STAMP 1)) (- (shift wcp:COUNTER 1) (+ wcp:COUNTER 1)))))

(defconstraint wcp:stamp-increments () (* (- (shift wcp:WORD_COMPARISON_STAMP 1) wcp:WORD_COMPARISON_STAMP) (- (shift wcp:WORD_COMPARISON_STAMP 1) (+ wcp:WORD_COMPARISON_STAMP 1))))

(defconstraint wcp:ct-upper-bond () (- (~ (- 16 wcp:COUNTER)) 1))

(defconstraint wcp:counter-reset () (ifnot (- (shift wcp:WORD_COMPARISON_STAMP 1) wcp:WORD_COMPARISON_STAMP) (shift wcp:COUNTER 1)))

(defconstraint wcp:setting-ct-max () (if (- wcp:ONE_LINE_INSTRUCTION 1) wcp:CT_MAX))

(defconstraint wcp:no-neg-if-small () (ifnot (- wcp:CT_MAX 15) (begin wcp:NEG_1 wcp:NEG_2)))

(defconstraint wcp:lastRow (:domain {-1}) (- wcp:COUNTER wcp:CT_MAX))

(defconstraint wcp:first-row (:domain {0}) wcp:WORD_COMPARISON_STAMP)
