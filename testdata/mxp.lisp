(defcolumns
  (mxp:SIZE_2_HI :u128)
  (mxp:MXP_TYPE_5 :u1)
  (mxp:BYTE_1 :u8)
  (mxp:ACC_Q :u136)
  (mxp:SIZE_2_LO :u128)
  (mxp:C_MEM_NEW :u64)
  (mxp:OFFSET_2_LO :u128)
  (mxp:BYTE_2 :u8)
  (mxp:C_MEM :u64)
  (mxp:MXPX :u1)
  (mxp:BYTE_Q :u8)
  (mxp:MXP_TYPE_3 :u1)
  (mxp:ACC_1 :u136)
  (mxp:COMP :u1)
  (mxp:WORDS_NEW :u64)
  (mxp:SIZE_1_LO :u128)
  (mxp:ACC_3 :u136)
  (mxp:QUAD_COST :u64)
  (mxp:GAS_MXP :u64)
  (mxp:ROOB :u1)
  (mxp:OFFSET_2_HI :u128)
  (mxp:BYTE_A :u8)
  (mxp:BYTE_QQ :u8)
  (mxp:CN :u64)
  (mxp:OFFSET_1_HI :u128)
  (mxp:MAX_OFFSET_1)
  (mxp:BYTE_W :u8)
  (mxp:BYTE_4 :u8)
  (mxp:BYTE_3 :u8)
  (mxp:STAMP :u32)
  (mxp:EXPANDS :u1)
  (mxp:MTNTOP :u1)
  (mxp:MXP_TYPE_1 :u1)
  (mxp:GWORD :u64)
  (mxp:LIN_COST :u64)
  (mxp:CT :u5)
  (mxp:MAX_OFFSET)
  (mxp:ACC_2 :u136)
  (mxp:BYTE_R :u8)
  (mxp:GBYTE :u64)
  (mxp:NOOP :u1)
  (mxp:WORDS :u64)
  (mxp:MAX_OFFSET_2)
  (mxp:ACC_A :u136)
  (mxp:ACC_W :u136)
  (mxp:OFFSET_1_LO :u128)
  (mxp:MXP_TYPE_4 :u1)
  (mxp:SIZE_1_HI :u128)
  (mxp:ACC_4 :u136)
  (mxp:INST :u8)
  (mxp:DEPLOYS :u1)
  (mxp:MXP_TYPE_2 :u1))

(defpermutation (mxp:CN_perm mxp:STAMP_perm mxp:C_MEM_perm mxp:C_MEM_NEW_perm mxp:WORDS_perm mxp:WORDS_NEW_perm) ((+ mxp:CN) (+ mxp:STAMP) (+ mxp:C_MEM) (+ mxp:C_MEM_NEW) (+ mxp:WORDS) (+ mxp:WORDS_NEW)))

(defconstraint mxp:counter-constancy () (begin (ifnot mxp:CT (- mxp:INST (shift mxp:INST -1))) (ifnot mxp:CT (- mxp:OFFSET_1_LO (shift mxp:OFFSET_1_LO -1))) (ifnot mxp:CT (- mxp:OFFSET_1_HI (shift mxp:OFFSET_1_HI -1))) (ifnot mxp:CT (- mxp:OFFSET_2_LO (shift mxp:OFFSET_2_LO -1))) (ifnot mxp:CT (- mxp:OFFSET_2_HI (shift mxp:OFFSET_2_HI -1))) (ifnot mxp:CT (- mxp:SIZE_1_LO (shift mxp:SIZE_1_LO -1))) (ifnot mxp:CT (- mxp:SIZE_1_HI (shift mxp:SIZE_1_HI -1))) (ifnot mxp:CT (- mxp:SIZE_2_LO (shift mxp:SIZE_2_LO -1))) (ifnot mxp:CT (- mxp:SIZE_2_HI (shift mxp:SIZE_2_HI -1))) (ifnot mxp:CT (- mxp:WORDS (shift mxp:WORDS -1))) (ifnot mxp:CT (- mxp:WORDS_NEW (shift mxp:WORDS_NEW -1))) (ifnot mxp:CT (- mxp:C_MEM (shift mxp:C_MEM -1))) (ifnot mxp:CT (- mxp:C_MEM_NEW (shift mxp:C_MEM_NEW -1))) (ifnot mxp:CT (- mxp:COMP (shift mxp:COMP -1))) (ifnot mxp:CT (- mxp:MXPX (shift mxp:MXPX -1))) (ifnot mxp:CT (- mxp:EXPANDS (shift mxp:EXPANDS -1))) (ifnot mxp:CT (- mxp:QUAD_COST (shift mxp:QUAD_COST -1))) (ifnot mxp:CT (- mxp:LIN_COST (shift mxp:LIN_COST -1))) (ifnot mxp:CT (- mxp:GAS_MXP (shift mxp:GAS_MXP -1)))))

(defconstraint mxp:byte-decompositions () (begin (if mxp:CT (- mxp:ACC_1 mxp:BYTE_1) (- mxp:ACC_1 (+ (* 256 (shift mxp:ACC_1 -1)) mxp:BYTE_1))) (if mxp:CT (- mxp:ACC_2 mxp:BYTE_2) (- mxp:ACC_2 (+ (* 256 (shift mxp:ACC_2 -1)) mxp:BYTE_2))) (if mxp:CT (- mxp:ACC_3 mxp:BYTE_3) (- mxp:ACC_3 (+ (* 256 (shift mxp:ACC_3 -1)) mxp:BYTE_3))) (if mxp:CT (- mxp:ACC_4 mxp:BYTE_4) (- mxp:ACC_4 (+ (* 256 (shift mxp:ACC_4 -1)) mxp:BYTE_4))) (if mxp:CT (- mxp:ACC_A mxp:BYTE_A) (- mxp:ACC_A (+ (* 256 (shift mxp:ACC_A -1)) mxp:BYTE_A))) (if mxp:CT (- mxp:ACC_W mxp:BYTE_W) (- mxp:ACC_W (+ (* 256 (shift mxp:ACC_W -1)) mxp:BYTE_W))) (if mxp:CT (- mxp:ACC_Q mxp:BYTE_Q) (- mxp:ACC_Q (+ (* 256 (shift mxp:ACC_Q -1)) mxp:BYTE_Q)))))

(defconstraint mxp:euclidean-division-of-square-of-accA () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX)) mxp:EXPANDS)) (begin (- (* mxp:ACC_A mxp:ACC_A) (+ (* 512 (+ mxp:ACC_Q (+ (* 4294967296 (shift mxp:BYTE_QQ -2)) (* 1099511627776 (shift mxp:BYTE_QQ -3))))) (+ (* 256 (shift mxp:BYTE_QQ -1)) mxp:BYTE_QQ))) (* (shift mxp:BYTE_QQ -1) (- 1 (shift mxp:BYTE_QQ -1))))))

(defconstraint mxp:setting-c-mem-new () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX)) mxp:EXPANDS)) (- mxp:C_MEM_NEW (+ (* 3 mxp:ACC_A) (+ mxp:ACC_Q (+ (* 4294967296 (shift mxp:BYTE_QQ -2)) (* 1099511627776 (shift mxp:BYTE_QQ -3))))))))

(defconstraint mxp:setting-roob-type-5 () (ifnot mxp:MXP_TYPE_5 (begin (ifnot mxp:SIZE_1_HI (- mxp:ROOB 1)) (ifnot mxp:SIZE_2_HI (- mxp:ROOB 1)) (ifnot (* mxp:OFFSET_1_HI mxp:SIZE_1_LO) (- mxp:ROOB 1)) (ifnot (* mxp:OFFSET_2_HI mxp:SIZE_2_LO) (- mxp:ROOB 1)) (if mxp:SIZE_1_HI (if mxp:SIZE_2_HI (if (* mxp:OFFSET_1_HI mxp:SIZE_1_LO) (if (* mxp:OFFSET_2_HI mxp:SIZE_2_LO) mxp:ROOB)))))))

(defconstraint mxp:setting-noop () (if mxp:ROOB (begin (ifnot (+ mxp:MXP_TYPE_1 mxp:MXP_TYPE_2 mxp:MXP_TYPE_3) (- mxp:NOOP mxp:MXP_TYPE_1)) (if (- mxp:MXP_TYPE_4 1) (- mxp:NOOP (- 1 (~ mxp:SIZE_1_LO)))) (if (- mxp:MXP_TYPE_5 1) (- mxp:NOOP (* (- 1 (~ mxp:SIZE_1_LO)) (- 1 (~ mxp:SIZE_2_LO))))))))

(defconstraint mxp:non-trivial-instruction-counter-cycle () (ifnot mxp:STAMP (ifnot (- 1 (+ mxp:ROOB mxp:NOOP)) (if mxp:MXPX (if (- mxp:CT 3) (- (shift mxp:STAMP 1) (+ mxp:STAMP 1)) (- (shift mxp:CT 1) (+ mxp:CT 1))) (if (- mxp:CT 16) (- (shift mxp:STAMP 1) (+ mxp:STAMP 1)) (- (shift mxp:CT 1) (+ mxp:CT 1)))))))

(defconstraint mxp:size-in-evm-words () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX))) (if (- mxp:MXP_TYPE_4 1) (begin (- mxp:SIZE_1_LO (- (* 32 mxp:ACC_W) mxp:BYTE_R)) (- (shift mxp:BYTE_R -1) (+ 224 mxp:BYTE_R))))))

(defconstraint mxp:comparing-max-offsets-1-and-2 () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX))) (- (+ mxp:ACC_3 (- 1 mxp:COMP)) (* (- mxp:MAX_OFFSET_1 mxp:MAX_OFFSET_2) (- (* 2 mxp:COMP) 1)))))

(defconstraint mxp:defining-accA () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX))) (begin (- (+ mxp:MAX_OFFSET 1) (- (* 32 mxp:ACC_A) (shift mxp:BYTE_R -2))) (- (shift mxp:BYTE_R -3) (+ 224 (shift mxp:BYTE_R -2))))))

(defconstraint mxp:setting-gas-mxp () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX))) (if (- mxp:INST 243) (- mxp:GAS_MXP (+ mxp:QUAD_COST (* mxp:DEPLOYS mxp:LIN_COST))) (- mxp:GAS_MXP (+ mxp:QUAD_COST mxp:LIN_COST)))))

(defconstraint mxp:mem-expansion-took-place () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX))) (- (+ mxp:ACC_4 mxp:EXPANDS) (* (- mxp:ACC_A mxp:WORDS) (- (* 2 mxp:EXPANDS) 1)))))

(defconstraint mxp:setting-quad-cost-and-lin-cost () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX))) (begin (- mxp:QUAD_COST (- mxp:C_MEM_NEW mxp:C_MEM)) (- mxp:LIN_COST (+ (* mxp:GBYTE mxp:SIZE_1_LO) (* mxp:GWORD mxp:ACC_W))))))

(defconstraint mxp:defining-max-offset () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX))) (- mxp:MAX_OFFSET (+ (* mxp:COMP mxp:MAX_OFFSET_1) (* (- 1 mxp:COMP) mxp:MAX_OFFSET_2)))))

(defconstraint mxp:max-offsets-1-and-2-type-5 () (ifnot (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (if (- mxp:MXP_TYPE_5 1) (begin (if mxp:SIZE_1_LO mxp:MAX_OFFSET_1 (- mxp:MAX_OFFSET_1 (+ mxp:OFFSET_1_LO (- mxp:SIZE_1_LO 1)))) (if mxp:SIZE_2_LO mxp:MAX_OFFSET_2 (- mxp:MAX_OFFSET_2 (+ mxp:OFFSET_2_LO (- mxp:SIZE_2_LO 1))))))))

(defconstraint mxp:binary-constraints () (begin (* mxp:ROOB (- 1 mxp:ROOB)) (* mxp:NOOP (- 1 mxp:NOOP)) (* mxp:MXPX (- 1 mxp:MXPX)) (* mxp:DEPLOYS (- 1 mxp:DEPLOYS)) (* mxp:COMP (- 1 mxp:COMP)) (* mxp:EXPANDS (- 1 mxp:EXPANDS))))

(defconstraint mxp:offsets-out-of-bounds () (ifnot (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (if (- mxp:MXPX 1) (if (- mxp:CT 16) (* (- (- mxp:MAX_OFFSET_1 4294967296) mxp:ACC_1) (- (- mxp:MAX_OFFSET_2 4294967296) mxp:ACC_2))))))

(defconstraint mxp:no-expansion () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX))) (if mxp:EXPANDS (begin (- mxp:WORDS_NEW mxp:WORDS) (- mxp:C_MEM_NEW mxp:C_MEM)))))

(defconstraint mxp:max-offsets-1-and-2-are-small () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX))) (begin (- mxp:ACC_1 mxp:MAX_OFFSET_1) (- mxp:ACC_2 mxp:MAX_OFFSET_2))))

(defconstraint mxp:setting-words-new () (ifnot (* (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (* (* (- 1 (~ (- mxp:CT 3))) (- 1 mxp:MXPX)) mxp:EXPANDS)) (- mxp:WORDS_NEW mxp:ACC_A)))

(defconstraint mxp:setting-roob-type-4 () (ifnot mxp:MXP_TYPE_4 (begin (ifnot mxp:SIZE_1_HI (- mxp:ROOB 1)) (ifnot (* mxp:OFFSET_1_HI mxp:SIZE_1_LO) (- mxp:ROOB 1)) (if mxp:SIZE_1_HI (if (* mxp:OFFSET_1_HI mxp:SIZE_1_LO) mxp:ROOB)))))

(defconstraint mxp:max-offsets-1-and-2-type-4 () (ifnot (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (if (- mxp:MXP_TYPE_4 1) (begin (- mxp:MAX_OFFSET_1 (+ mxp:OFFSET_1_LO (- mxp:SIZE_1_LO 1))) mxp:MAX_OFFSET_2))))

(defconstraint mxp:max-offsets-1-and-2-type-2 () (ifnot (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (if (- mxp:MXP_TYPE_2 1) (begin (- mxp:MAX_OFFSET_1 (+ mxp:OFFSET_1_LO 31)) mxp:MAX_OFFSET_2))))

(defconstraint mxp:consistency () (ifnot mxp:CN_perm (if (- (shift mxp:CN_perm -1) mxp:CN_perm) (ifnot (- (shift mxp:STAMP_perm -1) mxp:STAMP_perm) (begin (- mxp:WORDS_perm (shift mxp:WORDS_NEW_perm -1)) (- mxp:C_MEM_perm (shift mxp:C_MEM_NEW_perm -1)))) (begin mxp:WORDS_perm mxp:C_MEM_perm))))

(defconstraint mxp:type-flag-sum () (ifnot mxp:STAMP (- 1 (+ mxp:MXP_TYPE_1 (+ mxp:MXP_TYPE_2 (+ mxp:MXP_TYPE_3 (+ mxp:MXP_TYPE_5 mxp:MXP_TYPE_4)))))))

(defconstraint mxp:max-offsets-1-and-2-type-3 () (ifnot (* mxp:STAMP (- 1 mxp:NOOP mxp:ROOB)) (if (- mxp:MXP_TYPE_3 1) (begin (- mxp:MAX_OFFSET_1 mxp:OFFSET_1_LO) mxp:MAX_OFFSET_2))))

(defconstraint mxp:stamp-increment-when-roob-or-noop () (ifnot (+ mxp:ROOB mxp:NOOP) (begin (- (shift mxp:STAMP 1) (+ mxp:STAMP 1)) (- mxp:MXPX mxp:ROOB))))

(defconstraint mxp:final-row (:domain {-1}) (ifnot mxp:STAMP (if (+ mxp:ROOB mxp:NOOP) (- mxp:CT (if mxp:MXPX 3 16)))))

(defconstraint mxp:setting-roob-type-2-3 () (ifnot (+ mxp:MXP_TYPE_2 mxp:MXP_TYPE_3) (if mxp:OFFSET_1_HI mxp:ROOB (- mxp:ROOB 1))))

(defconstraint mxp:setting-mtntop () (if mxp:MXP_TYPE_4 mxp:MTNTOP (begin (if mxp:MXPX (if mxp:SIZE_1_LO mxp:MTNTOP (- mxp:MTNTOP 1)) mxp:MTNTOP))))

(defconstraint mxp:stamp-increments () (* (- (shift mxp:STAMP 1) mxp:STAMP) (- (shift mxp:STAMP 1) (+ mxp:STAMP 1))))

(defconstraint mxp:noop-consequences () (ifnot mxp:NOOP (begin mxp:QUAD_COST mxp:LIN_COST (- mxp:WORDS_NEW mxp:WORDS) (- mxp:C_MEM_NEW mxp:C_MEM))))

(defconstraint mxp:automatic-vanishing-when-padding () (if mxp:STAMP (begin (+ mxp:ROOB mxp:NOOP mxp:MXPX) mxp:CT mxp:INST)))

(defconstraint mxp:counter-reset () (ifnot (- (shift mxp:STAMP 1) mxp:STAMP) (shift mxp:CT 1)))

(defconstraint mxp:setting-roob-type-1 () (ifnot mxp:MXP_TYPE_1 mxp:ROOB))

(defconstraint mxp:noop-automatic-vanishing () (ifnot mxp:ROOB mxp:NOOP))

(defconstraint mxp:first-row (:domain {0}) mxp:STAMP)
