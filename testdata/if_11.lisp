(defpurefun ((eq! :@loob) x y) (- x y))
(defpurefun (if-eq x val then) (if (eq! x val) then))
;;
(defcolumns
  (CT :byte)
  (IS_SLT :binary@prove)
  (BITS :binary@prove)
  (NEG_2 :binary@prove))

(defconstraint bits-and-negs (:guard IS_SLT)
  (if-eq CT 15
	 (eq! NEG_2 (shift BITS (- 0 7)))))
