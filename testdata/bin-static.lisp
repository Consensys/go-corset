(module bin)

(defcolumns
  STAMP
  (ONE_LINE_INSTRUCTION :binary)
  (MLI :binary)
  (COUNTER :byte)
  (INST :byte)
  ARGUMENT_1_HI
  ARGUMENT_1_LO
  ARGUMENT_2_HI
  ARGUMENT_2_LO
  RESULT_HI
  RESULT_LO
  (IS_AND :binary)
  (IS_OR :binary)
  (IS_XOR :binary)
  (IS_NOT :binary)
  (IS_BYTE :binary)
  (IS_SIGNEXTEND :binary)
  (SMALL :binary)
  (BITS :binary)
  (BIT_B_4 :binary)
  (LOW_4 :byte)
  (NEG :binary)
  (BIT_1 :binary)
  (PIVOT :byte)
  (BYTE_1 :byte)
  (BYTE_2 :byte)
  (BYTE_3 :byte)
  (BYTE_4 :byte)
  (BYTE_5 :byte)
  (BYTE_6 :byte)
  ACC_1
  ACC_2
  ACC_3
  ACC_4
  ACC_5
  ACC_6
  ;; decoded bytes:
  (XXX_BYTE_HI :byte)
  (XXX_BYTE_LO :byte))


(defconstraint byte_decompositions () (begin (if COUNTER (- ACC_1 BYTE_1) (- ACC_1 (+ (* 256 (shift ACC_1 -1)) BYTE_1))) (if COUNTER (- ACC_2 BYTE_2) (- ACC_2 (+ (* 256 (shift ACC_2 -1)) BYTE_2))) (if COUNTER (- ACC_3 BYTE_3) (- ACC_3 (+ (* 256 (shift ACC_3 -1)) BYTE_3))) (if COUNTER (- ACC_4 BYTE_4) (- ACC_4 (+ (* 256 (shift ACC_4 -1)) BYTE_4))) (if COUNTER (- ACC_5 BYTE_5) (- ACC_5 (+ (* 256 (shift ACC_5 -1)) BYTE_5))) (if COUNTER (- ACC_6 BYTE_6) (- ACC_6 (+ (* 256 (shift ACC_6 -1)) BYTE_6)))))

(defconstraint bits-and-related () (if (- COUNTER 15) (begin (- PIVOT (+ (* 128 (shift BITS -15)) (* 64 (shift BITS -14)) (* 32 (shift BITS -13)) (* 16 (shift BITS -12)) (* 8 (shift BITS -11)) (* 4 (shift BITS -10)) (* 2 (shift BITS -9)) (shift BITS -8))) (- BYTE_2 (+ (* 128 (shift BITS -7)) (* 64 (shift BITS -6)) (* 32 (shift BITS -5)) (* 16 (shift BITS -4)) (* 8 (shift BITS -3)) (* 4 (shift BITS -2)) (* 2 (shift BITS -1)) BITS)) (- LOW_4 (+ (* 8 (shift BITS -3)) (* 4 (shift BITS -2)) (* 2 (shift BITS -1)) BITS)) (- BIT_B_4 (shift BITS -4)) (- NEG (shift BITS -15)))))

(defconstraint pivot () (if MLI 0 (begin (if IS_BYTE 0 (if LOW_4 (if COUNTER (if BIT_B_4 (- PIVOT BYTE_3) (- PIVOT BYTE_4))) (if (+ (shift BIT_1 -1) (- 1 BIT_1)) (if BIT_B_4 (- PIVOT BYTE_3) (- PIVOT BYTE_4))))) (if IS_SIGNEXTEND 0 (if (- LOW_4 15) (if COUNTER (if BIT_B_4 (- PIVOT BYTE_4) (- PIVOT BYTE_3))) (if (+ (shift BIT_1 -1) (- 1 BIT_1)) (if BIT_B_4 (- PIVOT BYTE_4) (- PIVOT BYTE_3))))))))

(defconstraint bit_1 () (begin (if (- IS_BYTE 1) (begin (if LOW_4 (- BIT_1 1) (if (- COUNTER 0) BIT_1 (if (- COUNTER LOW_4) (- BIT_1 (+ (shift BIT_1 -1) 1)) (- BIT_1 (shift BIT_1 -1))))))) (if (- IS_SIGNEXTEND 1) (begin (if (- 15 LOW_4) (- BIT_1 1) (if (- COUNTER 0) BIT_1 (if (- COUNTER (- 15 LOW_4)) (- BIT_1 (+ (shift BIT_1 -1) 1)) (- BIT_1 (shift BIT_1 -1)))))))))

(defconstraint binary_constraints () (begin (* IS_AND (- 1 IS_AND)) (* IS_OR (- 1 IS_OR)) (* IS_XOR (- 1 IS_XOR)) (* IS_NOT (- 1 IS_NOT)) (* IS_BYTE (- 1 IS_BYTE)) (* IS_SIGNEXTEND (- 1 IS_SIGNEXTEND)) (* SMALL (- 1 SMALL)) (* BITS (- 1 BITS)) (* NEG (- 1 NEG)) (* BIT_B_4 (- 1 BIT_B_4)) (* BIT_1 (- 1 BIT_1))))

(defconstraint counter-constancies () (begin (if COUNTER 0 (- ARGUMENT_1_HI (shift ARGUMENT_1_HI -1))) (if COUNTER 0 (- ARGUMENT_1_LO (shift ARGUMENT_1_LO -1))) (if COUNTER 0 (- ARGUMENT_2_HI (shift ARGUMENT_2_HI -1))) (if COUNTER 0 (- ARGUMENT_2_LO (shift ARGUMENT_2_LO -1))) (if COUNTER 0 (- RESULT_HI (shift RESULT_HI -1))) (if COUNTER 0 (- RESULT_LO (shift RESULT_LO -1))) (if COUNTER 0 (- INST (shift INST -1))) (if COUNTER 0 (- PIVOT (shift PIVOT -1))) (if COUNTER 0 (- BIT_B_4 (shift BIT_B_4 -1))) (if COUNTER 0 (- LOW_4 (shift LOW_4 -1))) (if COUNTER 0 (- NEG (shift NEG -1)))))

(defconstraint is-signextend-result () (if IS_SIGNEXTEND 0 (if (- ONE_LINE_INSTRUCTION 1) (begin (- RESULT_HI ARGUMENT_2_HI) (- RESULT_LO ARGUMENT_2_LO)) (if SMALL (begin (- RESULT_HI ARGUMENT_2_HI) (- RESULT_LO ARGUMENT_2_LO)) (begin (if BIT_B_4 (begin (- BYTE_5 (* NEG 255)) (if BIT_1 (- BYTE_6 (* NEG 255)) (- BYTE_6 BYTE_4))) (begin (if BIT_1 (- BYTE_5 (* NEG 255)) (- BYTE_5 BYTE_3)) (- RESULT_LO ARGUMENT_2_LO))))))))

(defconstraint small () (if (- COUNTER 15) (if ARGUMENT_1_HI (if (- ARGUMENT_1_LO (+ (* 16 (shift BITS -4)) (* 8 (shift BITS -3)) (* 4 (shift BITS -2)) (* 2 (shift BITS -1)) BITS)) (- SMALL 1) SMALL))))

(defconstraint inst-to-flag () (- INST (+ (* IS_AND 22) (* IS_OR 23) (* IS_XOR 24) (* IS_NOT 25) (* IS_BYTE 26) (* IS_SIGNEXTEND 11))))

(defconstraint target-constraints () (if (- COUNTER 15) (begin (- ACC_1 ARGUMENT_1_HI) (- ACC_2 ARGUMENT_1_LO) (- ACC_3 ARGUMENT_2_HI) (- ACC_4 ARGUMENT_2_LO) (- ACC_5 RESULT_HI) (- ACC_6 RESULT_LO))))

(defconstraint mli-incrementation () (if MLI 0 (if (- COUNTER 15) (- (shift STAMP 1) (+ STAMP 1)) (- (shift COUNTER 1) (+ COUNTER 1)))))

(defconstraint stamp-increments () (* (- (shift STAMP 1) (+ STAMP 0)) (- (shift STAMP 1) (+ STAMP 1))))

(defconstraint is-byte-result () (if IS_BYTE 0 (if (- ONE_LINE_INSTRUCTION 1) (begin RESULT_HI RESULT_LO) (begin RESULT_HI (- RESULT_LO (* SMALL PIVOT))))))

(defconstraint no-bin-no-flag () (if STAMP (+ IS_AND IS_OR IS_XOR IS_NOT IS_BYTE IS_SIGNEXTEND) (- (+ IS_AND IS_OR IS_XOR IS_NOT IS_BYTE IS_SIGNEXTEND) 1)))

(defconstraint set-oli-mli () (if (+ IS_BYTE IS_SIGNEXTEND) ONE_LINE_INSTRUCTION (if ARGUMENT_1_HI ONE_LINE_INSTRUCTION (- ONE_LINE_INSTRUCTION 1))))

(defconstraint result-via-deflookup () (if (+ IS_AND IS_OR IS_XOR IS_NOT) 0 (begin (- BYTE_5 XXX_BYTE_HI) (- BYTE_6 XXX_BYTE_LO))))

(defconstraint oli-incrementation () (if ONE_LINE_INSTRUCTION 0 (- (shift STAMP 1) (+ STAMP 1))))

(defconstraint last-row (:domain {-1}) (if (- MLI 1) (- COUNTER 15)))

(defconstraint oli-mli-exclusivity () (- (+ ONE_LINE_INSTRUCTION MLI) (+ IS_AND IS_OR IS_XOR IS_NOT IS_BYTE IS_SIGNEXTEND)))

(defconstraint countereset () (if (- (shift STAMP 1) STAMP) 0 (shift COUNTER 1)))

(defconstraint first-row (:domain {0}) STAMP)
