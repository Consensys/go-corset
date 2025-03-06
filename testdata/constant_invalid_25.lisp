;;error:13:58-64:not permitted in const context
(defpurefun ((eq! :@loob) x y) (- x y))

(defcolumns
  (C :byte)
  (L :binary)
  (B :binary)
  (N :binary))

;; opcode values
(defconst
  (LLARGE :extern)                                    16
  (LLARGEMO :extern)                                  (- LLARGE 1))

(defconstraint bits-and-negs (:guard L)
  (if (eq! C LLARGEMO)
      (eq! N
	   (shift B (- 0 LLARGEMO)))))
