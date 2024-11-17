(module mxp)

(defcolumns
	(STAMP  :i32)
	(CN     :i64)
	(CT     :i5)
	(ROOB	:binary@prove)
	(NOOP	:binary@prove)
	(MXPX	:binary@prove)
	(INST   :byte)
	(MXP_TYPE_1 :binary@prove)
	(MXP_TYPE_2 :binary@prove)
	(MXP_TYPE_3 :binary@prove)
	(MXP_TYPE_4 :binary@prove)
	(MXP_TYPE_5 :binary@prove)
        (GBYTE :i64)
        (GWORD :i64)
        (DEPLOYS :binary@prove)
        (OFFSET_1_LO :i128)
        (OFFSET_2_LO :i128)
        (OFFSET_1_HI :i128)
        (OFFSET_2_HI :i128)
        (SIZE_1_LO :i128)
        (SIZE_2_LO :i128)
        (SIZE_1_HI :i128)
        (SIZE_2_HI :i128)
        (MAX_OFFSET_1 :i128)
        (MAX_OFFSET_2 :i128)
        (MAX_OFFSET   :i128)
        (COMP :binary@prove)
	(BYTE_1 :byte@prove)
	(BYTE_2 :byte@prove)
	(BYTE_3 :byte@prove)
	(BYTE_4 :byte@prove)
	(BYTE_A	:byte@prove)
	(BYTE_W	:byte@prove)
	(BYTE_Q	:byte@prove)
	(ACC_1 :i136)
	(ACC_2 :i136)
	(ACC_3 :i136)
	(ACC_4 :i136)
	(ACC_A :i136)
	(ACC_W :i136)
	(ACC_Q :i136)
        (BYTE_QQ :byte@prove)
        (BYTE_R :byte@prove)
        (WORDS :i64)
        (WORDS_NEW :i64)
        (C_MEM :i64)
        (C_MEM_NEW :i64)
        (QUAD_COST :i64)
        (LIN_COST :i64)
        (GAS_MXP :i64)
        (EXPANDS :binary@prove)
        (MTNTOP :binary@prove))

(defpermutation (CN_perm STAMP_perm C_MEM_perm C_MEM_NEW_perm WORDS_perm WORDS_NEW_perm) ((+ CN) (+ STAMP) (+ C_MEM) (+ C_MEM_NEW) (+ WORDS) (+ WORDS_NEW)))

(defconstraint counter-constancy () (begin (if CT 0 (- INST (shift INST -1))) (if CT 0 (- OFFSET_1_LO (shift OFFSET_1_LO -1))) (if CT 0 (- OFFSET_1_HI (shift OFFSET_1_HI -1))) (if CT 0 (- OFFSET_2_LO (shift OFFSET_2_LO -1))) (if CT 0 (- OFFSET_2_HI (shift OFFSET_2_HI -1))) (if CT 0 (- SIZE_1_LO (shift SIZE_1_LO -1))) (if CT 0 (- SIZE_1_HI (shift SIZE_1_HI -1))) (if CT 0 (- SIZE_2_LO (shift SIZE_2_LO -1))) (if CT 0 (- SIZE_2_HI (shift SIZE_2_HI -1))) (if CT 0 (- WORDS (shift WORDS -1))) (if CT 0 (- WORDS_NEW (shift WORDS_NEW -1))) (if CT 0 (- C_MEM (shift C_MEM -1))) (if CT 0 (- C_MEM_NEW (shift C_MEM_NEW -1))) (if CT 0 (- COMP (shift COMP -1))) (if CT 0 (- MXPX (shift MXPX -1))) (if CT 0 (- EXPANDS (shift EXPANDS -1))) (if CT 0 (- QUAD_COST (shift QUAD_COST -1))) (if CT 0 (- LIN_COST (shift LIN_COST -1))) (if CT 0 (- GAS_MXP (shift GAS_MXP -1)))))

(defconstraint byte-decompositions () (begin (if CT (- ACC_1 BYTE_1) (- ACC_1 (+ (* 256 (shift ACC_1 -1)) BYTE_1))) (if CT (- ACC_2 BYTE_2) (- ACC_2 (+ (* 256 (shift ACC_2 -1)) BYTE_2))) (if CT (- ACC_3 BYTE_3) (- ACC_3 (+ (* 256 (shift ACC_3 -1)) BYTE_3))) (if CT (- ACC_4 BYTE_4) (- ACC_4 (+ (* 256 (shift ACC_4 -1)) BYTE_4))) (if CT (- ACC_A BYTE_A) (- ACC_A (+ (* 256 (shift ACC_A -1)) BYTE_A))) (if CT (- ACC_W BYTE_W) (- ACC_W (+ (* 256 (shift ACC_W -1)) BYTE_W))) (if CT (- ACC_Q BYTE_Q) (- ACC_Q (+ (* 256 (shift ACC_Q -1)) BYTE_Q)))))

(defconstraint euclidean-division-of-square-of-accA () (if (* (* STAMP (- 1 NOOP ROOB)) (* (* (- 1 (~ (- CT 3))) (- 1 MXPX)) EXPANDS)) 0 (begin (- (* ACC_A ACC_A) (+ (* 512 (+ ACC_Q (+ (* 4294967296 (shift BYTE_QQ -2)) (* 1099511627776 (shift BYTE_QQ -3))))) (+ (* 256 (shift BYTE_QQ -1)) BYTE_QQ))) (* (shift BYTE_QQ -1) (- 1 (shift BYTE_QQ -1))))))

(defconstraint setting-c-mem-new () (if (* (* STAMP (- 1 NOOP ROOB)) (* (* (- 1 (~ (- CT 3))) (- 1 MXPX)) EXPANDS)) 0 (- C_MEM_NEW (+ (* 3 ACC_A) (+ ACC_Q (+ (* 4294967296 (shift BYTE_QQ -2)) (* 1099511627776 (shift BYTE_QQ -3))))))))

(defconstraint setting-roob-type-5 () (if MXP_TYPE_5 0 (begin (if SIZE_1_HI 0 (- ROOB 1)) (if SIZE_2_HI 0 (- ROOB 1)) (if (* OFFSET_1_HI SIZE_1_LO) 0 (- ROOB 1)) (if (* OFFSET_2_HI SIZE_2_LO) 0 (- ROOB 1)) (if SIZE_1_HI (if SIZE_2_HI (if (* OFFSET_1_HI SIZE_1_LO) (if (* OFFSET_2_HI SIZE_2_LO) ROOB)))))))

(defconstraint setting-noop () (if ROOB (begin (if (+ MXP_TYPE_1 MXP_TYPE_2 MXP_TYPE_3) 0 (- NOOP MXP_TYPE_1)) (if (- MXP_TYPE_4 1) (- NOOP (- 1 (~ SIZE_1_LO)))) (if (- MXP_TYPE_5 1) (- NOOP (* (- 1 (~ SIZE_1_LO)) (- 1 (~ SIZE_2_LO))))))))

(defconstraint non-trivial-instruction-counter-cycle () (if STAMP 0 (if (- 1 (+ ROOB NOOP)) 0 (if MXPX (if (- CT 3) (- (shift STAMP 1) (+ STAMP 1)) (- (shift CT 1) (+ CT 1))) (if (- CT 16) (- (shift STAMP 1) (+ STAMP 1)) (- (shift CT 1) (+ CT 1)))))))

(defconstraint size-in-evm-words () (if (* (* STAMP (- 1 NOOP ROOB)) (* (- 1 (~ (- CT 3))) (- 1 MXPX))) 0 (if (- MXP_TYPE_4 1) (begin (- SIZE_1_LO (- (* 32 ACC_W) BYTE_R)) (- (shift BYTE_R -1) (+ 224 BYTE_R))))))

(defconstraint comparing-max-offsets-1-and-2 () (if (* (* STAMP (- 1 NOOP ROOB)) (* (- 1 (~ (- CT 3))) (- 1 MXPX))) 0 (- (+ ACC_3 (- 1 COMP)) (* (- MAX_OFFSET_1 MAX_OFFSET_2) (- (* 2 COMP) 1)))))

(defconstraint defining-accA () (if (* (* STAMP (- 1 NOOP ROOB)) (* (- 1 (~ (- CT 3))) (- 1 MXPX))) 0 (begin (- (+ MAX_OFFSET 1) (- (* 32 ACC_A) (shift BYTE_R -2))) (- (shift BYTE_R -3) (+ 224 (shift BYTE_R -2))))))

(defconstraint setting-gas-mxp () (if (* (* STAMP (- 1 NOOP ROOB)) (* (- 1 (~ (- CT 3))) (- 1 MXPX))) 0 (if (- INST 243) (- GAS_MXP (+ QUAD_COST (* DEPLOYS LIN_COST))) (- GAS_MXP (+ QUAD_COST LIN_COST)))))

(defconstraint mem-expansion-took-place () (if (* (* STAMP (- 1 NOOP ROOB)) (* (- 1 (~ (- CT 3))) (- 1 MXPX))) 0 (- (+ ACC_4 EXPANDS) (* (- ACC_A WORDS) (- (* 2 EXPANDS) 1)))))

(defconstraint setting-quad-cost-and-lin-cost () (if (* (* STAMP (- 1 NOOP ROOB)) (* (- 1 (~ (- CT 3))) (- 1 MXPX))) 0 (begin (- QUAD_COST (- C_MEM_NEW C_MEM)) (- LIN_COST (+ (* GBYTE SIZE_1_LO) (* GWORD ACC_W))))))

(defconstraint defining-max-offset () (if (* (* STAMP (- 1 NOOP ROOB)) (* (- 1 (~ (- CT 3))) (- 1 MXPX))) 0 (- MAX_OFFSET (+ (* COMP MAX_OFFSET_1) (* (- 1 COMP) MAX_OFFSET_2)))))

(defconstraint max-offsets-1-and-2-type-5 () (if (* STAMP (- 1 NOOP ROOB)) 0 (if (- MXP_TYPE_5 1) (begin (if SIZE_1_LO MAX_OFFSET_1 (- MAX_OFFSET_1 (+ OFFSET_1_LO (- SIZE_1_LO 1)))) (if SIZE_2_LO MAX_OFFSET_2 (- MAX_OFFSET_2 (+ OFFSET_2_LO (- SIZE_2_LO 1))))))))

(defconstraint binary-constraints () (begin (* ROOB (- 1 ROOB)) (* NOOP (- 1 NOOP)) (* MXPX (- 1 MXPX)) (* DEPLOYS (- 1 DEPLOYS)) (* COMP (- 1 COMP)) (* EXPANDS (- 1 EXPANDS))))

(defconstraint offsets-out-of-bounds () (if (* STAMP (- 1 NOOP ROOB)) 0 (if (- MXPX 1) (if (- CT 16) (* (- (- MAX_OFFSET_1 4294967296) ACC_1) (- (- MAX_OFFSET_2 4294967296) ACC_2))))))

(defconstraint no-expansion () (if (* (* STAMP (- 1 NOOP ROOB)) (* (- 1 (~ (- CT 3))) (- 1 MXPX))) 0 (if EXPANDS (begin (- WORDS_NEW WORDS) (- C_MEM_NEW C_MEM)))))

(defconstraint max-offsets-1-and-2-are-small () (if (* (* STAMP (- 1 NOOP ROOB)) (* (- 1 (~ (- CT 3))) (- 1 MXPX))) 0 (begin (- ACC_1 MAX_OFFSET_1) (- ACC_2 MAX_OFFSET_2))))

(defconstraint setting-words-new () (if (* (* STAMP (- 1 NOOP ROOB)) (* (* (- 1 (~ (- CT 3))) (- 1 MXPX)) EXPANDS)) 0 (- WORDS_NEW ACC_A)))

(defconstraint setting-roob-type-4 () (if MXP_TYPE_4 0 (begin (if SIZE_1_HI 0 (- ROOB 1)) (if (* OFFSET_1_HI SIZE_1_LO) 0 (- ROOB 1)) (if SIZE_1_HI (if (* OFFSET_1_HI SIZE_1_LO) ROOB)))))

(defconstraint max-offsets-1-and-2-type-4 () (if (* STAMP (- 1 NOOP ROOB)) 0 (if (- MXP_TYPE_4 1) (begin (- MAX_OFFSET_1 (+ OFFSET_1_LO (- SIZE_1_LO 1))) MAX_OFFSET_2))))

(defconstraint max-offsets-1-and-2-type-2 () (if (* STAMP (- 1 NOOP ROOB)) 0 (if (- MXP_TYPE_2 1) (begin (- MAX_OFFSET_1 (+ OFFSET_1_LO 31)) MAX_OFFSET_2))))

(defconstraint consistency () (if CN_perm 0 (if (- (shift CN_perm -1) CN_perm) (if (- (shift STAMP_perm -1) STAMP_perm) 0 (begin (- WORDS_perm (shift WORDS_NEW_perm -1)) (- C_MEM_perm (shift C_MEM_NEW_perm -1)))) (begin WORDS_perm C_MEM_perm))))

(defconstraint type-flag-sum () (if STAMP 0 (- 1 (+ MXP_TYPE_1 (+ MXP_TYPE_2 (+ MXP_TYPE_3 (+ MXP_TYPE_5 MXP_TYPE_4)))))))

(defconstraint max-offsets-1-and-2-type-3 () (if (* STAMP (- 1 NOOP ROOB)) 0 (if (- MXP_TYPE_3 1) (begin (- MAX_OFFSET_1 OFFSET_1_LO) MAX_OFFSET_2))))

(defconstraint stamp-increment-when-roob-or-noop () (if (+ ROOB NOOP) 0 (begin (- (shift STAMP 1) (+ STAMP 1)) (- MXPX ROOB))))

(defconstraint final-row (:domain {-1}) (if STAMP 0 (if (+ ROOB NOOP) (- CT (if MXPX 3 16)))))

(defconstraint setting-roob-type-2-3 () (if (+ MXP_TYPE_2 MXP_TYPE_3) 0 (if OFFSET_1_HI ROOB (- ROOB 1))))

(defconstraint setting-mtntop () (if MXP_TYPE_4 MTNTOP (begin (if MXPX (if SIZE_1_LO MTNTOP (- MTNTOP 1)) MTNTOP))))

(defconstraint stamp-increments () (* (- (shift STAMP 1) STAMP) (- (shift STAMP 1) (+ STAMP 1))))

(defconstraint noop-consequences () (if NOOP 0 (begin QUAD_COST LIN_COST (- WORDS_NEW WORDS) (- C_MEM_NEW C_MEM))))

(defconstraint automatic-vanishing-when-padding () (if STAMP (begin (+ ROOB NOOP MXPX) CT INST)))

(defconstraint counter-reset () (if (- (shift STAMP 1) STAMP) 0 (shift CT 1)))

(defconstraint setting-roob-type-1 () (if MXP_TYPE_1 0 ROOB))

(defconstraint noop-automatic-vanishing () (if ROOB 0 NOOP))

(defconstraint first-row (:domain {0}) STAMP)
