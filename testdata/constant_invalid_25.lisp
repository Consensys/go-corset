;;error:12:58-64:not permitted in const context

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
  (if (== C LLARGEMO)
      (== N
	   (shift B (- 0 LLARGEMO)))))
