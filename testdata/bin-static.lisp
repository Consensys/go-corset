(defcolumns
  (bin:BYTE_5 :u8)
  (bin:ARGUMENT_1_LO)
  (bin:ACC_5)
  (bin:PIVOT :u8)
  (bin:ARGUMENT_1_HI)
  (bin:SMALL :u1)
  (bin:IS_OR :u1)
  (bin:BYTE_4 :u8)
  (bin:XXX_BYTE_HI :u8)
  (bin:ACC_3)
  (bin:XXX_BYTE_LO :u8)
  (bin:MLI :u1)
  (bin:LOW_4 :u8)
  (bin:BYTE_3 :u8)
  (bin:BIT_B_4 :u1)
  (bin:RESULT_LO)
  (bin:ACC_1)
  (bin:IS_SIGNEXTEND :u1)
  (bin:IS_NOT :u1)
  (bin:ACC_6)
  (bin:BITS :u1)
  (bin:BIT_1 :u1)
  (bin:ARGUMENT_2_LO)
  (bin:IS_BYTE :u1)
  (bin:ACC_4)
  (bin:ONE_LINE_INSTRUCTION :u1)
  (bin:COUNTER :u8)
  (bin:ACC_2)
  (bin:RESULT_HI)
  (bin:NEG :u1)
  (bin:BYTE_2 :u8)
  (bin:INST :u8)
  (bin:ARGUMENT_2_HI)
  (bin:BYTE_1 :u8)
  (bin:IS_AND :u1)
  (bin:BYTE_6 :u8)
  (bin:STAMP)
  (bin:IS_XOR :u1))

(defconstraint bin:byte_decompositions () (begin (if bin:COUNTER (- bin:ACC_1 bin:BYTE_1) (- bin:ACC_1 (+ (* 256 (shift bin:ACC_1 -1)) bin:BYTE_1))) (if bin:COUNTER (- bin:ACC_2 bin:BYTE_2) (- bin:ACC_2 (+ (* 256 (shift bin:ACC_2 -1)) bin:BYTE_2))) (if bin:COUNTER (- bin:ACC_3 bin:BYTE_3) (- bin:ACC_3 (+ (* 256 (shift bin:ACC_3 -1)) bin:BYTE_3))) (if bin:COUNTER (- bin:ACC_4 bin:BYTE_4) (- bin:ACC_4 (+ (* 256 (shift bin:ACC_4 -1)) bin:BYTE_4))) (if bin:COUNTER (- bin:ACC_5 bin:BYTE_5) (- bin:ACC_5 (+ (* 256 (shift bin:ACC_5 -1)) bin:BYTE_5))) (if bin:COUNTER (- bin:ACC_6 bin:BYTE_6) (- bin:ACC_6 (+ (* 256 (shift bin:ACC_6 -1)) bin:BYTE_6)))))

(defconstraint bin:bits-and-related () (if (- bin:COUNTER 15) (begin (- bin:PIVOT (+ (* 128 (shift bin:BITS -15)) (* 64 (shift bin:BITS -14)) (* 32 (shift bin:BITS -13)) (* 16 (shift bin:BITS -12)) (* 8 (shift bin:BITS -11)) (* 4 (shift bin:BITS -10)) (* 2 (shift bin:BITS -9)) (shift bin:BITS -8))) (- bin:BYTE_2 (+ (* 128 (shift bin:BITS -7)) (* 64 (shift bin:BITS -6)) (* 32 (shift bin:BITS -5)) (* 16 (shift bin:BITS -4)) (* 8 (shift bin:BITS -3)) (* 4 (shift bin:BITS -2)) (* 2 (shift bin:BITS -1)) bin:BITS)) (- bin:LOW_4 (+ (* 8 (shift bin:BITS -3)) (* 4 (shift bin:BITS -2)) (* 2 (shift bin:BITS -1)) bin:BITS)) (- bin:BIT_B_4 (shift bin:BITS -4)) (- bin:NEG (shift bin:BITS -15)))))

(defconstraint bin:pivot () (ifnot bin:MLI (begin (ifnot bin:IS_BYTE (if bin:LOW_4 (if bin:COUNTER (if bin:BIT_B_4 (- bin:PIVOT bin:BYTE_3) (- bin:PIVOT bin:BYTE_4))) (if (+ (shift bin:BIT_1 -1) (- 1 bin:BIT_1)) (if bin:BIT_B_4 (- bin:PIVOT bin:BYTE_3) (- bin:PIVOT bin:BYTE_4))))) (ifnot bin:IS_SIGNEXTEND (if (- bin:LOW_4 15) (if bin:COUNTER (if bin:BIT_B_4 (- bin:PIVOT bin:BYTE_4) (- bin:PIVOT bin:BYTE_3))) (if (+ (shift bin:BIT_1 -1) (- 1 bin:BIT_1)) (if bin:BIT_B_4 (- bin:PIVOT bin:BYTE_4) (- bin:PIVOT bin:BYTE_3))))))))

(defconstraint bin:bit_1 () (begin (if (- bin:IS_BYTE 1) (begin (if bin:LOW_4 (- bin:BIT_1 1) (if (- bin:COUNTER 0) bin:BIT_1 (if (- bin:COUNTER bin:LOW_4) (- bin:BIT_1 (+ (shift bin:BIT_1 -1) 1)) (- bin:BIT_1 (shift bin:BIT_1 -1))))))) (if (- bin:IS_SIGNEXTEND 1) (begin (if (- 15 bin:LOW_4) (- bin:BIT_1 1) (if (- bin:COUNTER 0) bin:BIT_1 (if (- bin:COUNTER (- 15 bin:LOW_4)) (- bin:BIT_1 (+ (shift bin:BIT_1 -1) 1)) (- bin:BIT_1 (shift bin:BIT_1 -1)))))))))

(defconstraint bin:binary_constraints () (begin (* bin:IS_AND (- 1 bin:IS_AND)) (* bin:IS_OR (- 1 bin:IS_OR)) (* bin:IS_XOR (- 1 bin:IS_XOR)) (* bin:IS_NOT (- 1 bin:IS_NOT)) (* bin:IS_BYTE (- 1 bin:IS_BYTE)) (* bin:IS_SIGNEXTEND (- 1 bin:IS_SIGNEXTEND)) (* bin:SMALL (- 1 bin:SMALL)) (* bin:BITS (- 1 bin:BITS)) (* bin:NEG (- 1 bin:NEG)) (* bin:BIT_B_4 (- 1 bin:BIT_B_4)) (* bin:BIT_1 (- 1 bin:BIT_1))))

(defconstraint bin:counter-constancies () (begin (ifnot bin:COUNTER (- bin:ARGUMENT_1_HI (shift bin:ARGUMENT_1_HI -1))) (ifnot bin:COUNTER (- bin:ARGUMENT_1_LO (shift bin:ARGUMENT_1_LO -1))) (ifnot bin:COUNTER (- bin:ARGUMENT_2_HI (shift bin:ARGUMENT_2_HI -1))) (ifnot bin:COUNTER (- bin:ARGUMENT_2_LO (shift bin:ARGUMENT_2_LO -1))) (ifnot bin:COUNTER (- bin:RESULT_HI (shift bin:RESULT_HI -1))) (ifnot bin:COUNTER (- bin:RESULT_LO (shift bin:RESULT_LO -1))) (ifnot bin:COUNTER (- bin:INST (shift bin:INST -1))) (ifnot bin:COUNTER (- bin:PIVOT (shift bin:PIVOT -1))) (ifnot bin:COUNTER (- bin:BIT_B_4 (shift bin:BIT_B_4 -1))) (ifnot bin:COUNTER (- bin:LOW_4 (shift bin:LOW_4 -1))) (ifnot bin:COUNTER (- bin:NEG (shift bin:NEG -1)))))

(defconstraint bin:is-signextend-result () (ifnot bin:IS_SIGNEXTEND (if (- bin:ONE_LINE_INSTRUCTION 1) (begin (- bin:RESULT_HI bin:ARGUMENT_2_HI) (- bin:RESULT_LO bin:ARGUMENT_2_LO)) (if bin:SMALL (begin (- bin:RESULT_HI bin:ARGUMENT_2_HI) (- bin:RESULT_LO bin:ARGUMENT_2_LO)) (begin (if bin:BIT_B_4 (begin (- bin:BYTE_5 (* bin:NEG 255)) (if bin:BIT_1 (- bin:BYTE_6 (* bin:NEG 255)) (- bin:BYTE_6 bin:BYTE_4))) (begin (if bin:BIT_1 (- bin:BYTE_5 (* bin:NEG 255)) (- bin:BYTE_5 bin:BYTE_3)) (- bin:RESULT_LO bin:ARGUMENT_2_LO))))))))

(defconstraint bin:small () (if (- bin:COUNTER 15) (if bin:ARGUMENT_1_HI (if (- bin:ARGUMENT_1_LO (+ (* 16 (shift bin:BITS -4)) (* 8 (shift bin:BITS -3)) (* 4 (shift bin:BITS -2)) (* 2 (shift bin:BITS -1)) bin:BITS)) (- bin:SMALL 1) bin:SMALL))))

(defconstraint bin:inst-to-flag () (- bin:INST (+ (* bin:IS_AND 22) (* bin:IS_OR 23) (* bin:IS_XOR 24) (* bin:IS_NOT 25) (* bin:IS_BYTE 26) (* bin:IS_SIGNEXTEND 11))))

(defconstraint bin:target-constraints () (if (- bin:COUNTER 15) (begin (- bin:ACC_1 bin:ARGUMENT_1_HI) (- bin:ACC_2 bin:ARGUMENT_1_LO) (- bin:ACC_3 bin:ARGUMENT_2_HI) (- bin:ACC_4 bin:ARGUMENT_2_LO) (- bin:ACC_5 bin:RESULT_HI) (- bin:ACC_6 bin:RESULT_LO))))

(defconstraint bin:mli-incrementation () (ifnot bin:MLI (if (- bin:COUNTER 15) (- (shift bin:STAMP 1) (+ bin:STAMP 1)) (- (shift bin:COUNTER 1) (+ bin:COUNTER 1)))))

(defconstraint bin:stamp-increments () (* (- (shift bin:STAMP 1) (+ bin:STAMP 0)) (- (shift bin:STAMP 1) (+ bin:STAMP 1))))

(defconstraint bin:is-byte-result () (ifnot bin:IS_BYTE (if (- bin:ONE_LINE_INSTRUCTION 1) (begin bin:RESULT_HI bin:RESULT_LO) (begin bin:RESULT_HI (- bin:RESULT_LO (* bin:SMALL bin:PIVOT))))))

(defconstraint bin:no-bin-no-flag () (if bin:STAMP (+ bin:IS_AND bin:IS_OR bin:IS_XOR bin:IS_NOT bin:IS_BYTE bin:IS_SIGNEXTEND) (- (+ bin:IS_AND bin:IS_OR bin:IS_XOR bin:IS_NOT bin:IS_BYTE bin:IS_SIGNEXTEND) 1)))

(defconstraint bin:set-oli-mli () (if (+ bin:IS_BYTE bin:IS_SIGNEXTEND) bin:ONE_LINE_INSTRUCTION (if bin:ARGUMENT_1_HI bin:ONE_LINE_INSTRUCTION (- bin:ONE_LINE_INSTRUCTION 1))))

(defconstraint bin:result-via-lookup () (ifnot (+ bin:IS_AND bin:IS_OR bin:IS_XOR bin:IS_NOT) (begin (- bin:BYTE_5 bin:XXX_BYTE_HI) (- bin:BYTE_6 bin:XXX_BYTE_LO))))

(defconstraint bin:oli-incrementation () (ifnot bin:ONE_LINE_INSTRUCTION (- (shift bin:STAMP 1) (+ bin:STAMP 1))))

(defconstraint bin:last-row (:domain {-1}) (if (- bin:MLI 1) (- bin:COUNTER 15)))

(defconstraint bin:oli-mli-exclusivity () (- (+ bin:ONE_LINE_INSTRUCTION bin:MLI) (+ bin:IS_AND bin:IS_OR bin:IS_XOR bin:IS_NOT bin:IS_BYTE bin:IS_SIGNEXTEND)))

(defconstraint bin:countereset () (ifnot (- (shift bin:STAMP 1) bin:STAMP) (shift bin:COUNTER 1)))

(defconstraint bin:first-row (:domain {0}) bin:STAMP)
