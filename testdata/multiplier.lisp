;; Shift & Add Multiplier
;;
;;     |x|y|
;;     |a|b|
;; ---------
;; | | |*|*| +b.y
;; | |*|*| | +b.x
;; | |*|*| | +a.y
;; |*|*| | | +a.x
(defcolumns
  (ST :i32)
  (CT :i4)
  ;; nibbles
  (ARG1 :i4@prove :array [0:1])
  (ARG2 :i4@prove :array [0:1])
  (RES :i4@prove :array [0:3]))

;; NOTE: this is not really a finished example, since it doesn't
;; actually use carry lines (yet).  These would be needed if small
;; fields are used to prevent overflow.

;; ===================================================================
;; Control
;; ===================================================================

;; In the first row, ST is always zero.  This allows for an
;; arbitrary amount of padding at the beginning which has no function.
(defconstraint first (:domain {0}) (== ST 0))

;; In the last row of a valid frame, the counter must have its max
;; value.  This ensures that all non-padding frames are complete.
(defconstraint last (:domain {-1} :guard ST)
  ;; CT[$] == 3
  (== CT 3))

;; ST either remains constant, or increments by one.
(defconstraint increment ()
  (or!
   ;; ST[k] == ST[k+1]
   (== ST (shift ST 1))
   ;; Or, ST[k]+1 == ST[k+1]
   (== (+ 1 ST) (next ST))))

;; If ST changes, counter resets to zero.
(defconstraint reset ()
  (or!
   ;; ST[k] == ST[k+1]
   (== ST (shift ST 1))
   ;; Or, CT[k+1] == 0
   (== (next CT) 0)))

;; Increment or reset counter
(defconstraint heartbeat (:guard ST)
  ;; If CT[k] == 3
  (if (== CT 3)
      ;; Then, CT[k+1] == 0
      (== (next CT) 0)
      ;; Else, CT[k]+1 == CT[k+1]
      (== (+ 1 CT) (next CT))))

;; ===================================================================
;; Multipilier
;; ===================================================================

(defconstraint line_1 (:guard ST)
  (if (== CT 0)
      (== (RES) (* [ARG1 0] [ARG2 0]))))

(defconstraint line_2 (:guard ST)
  (if (== CT 1)
      (== (RES) (+ (prev (RES)) (* 16 [ARG1 0] [ARG2 1])))))

(defconstraint line_3 (:guard ST)
  (if (== CT 2)
      (== (RES) (+ (prev (RES)) (* 16 [ARG1 1] [ARG2 0])))))

(defconstraint line_4 (:guard ST)
  (if (== CT 3)
      (== (RES) (+ (prev (RES)) (* 256 [ARG1 1] [ARG2 1])))))

;; ===================================================================
;; Helpers
;; ===================================================================

(defun (RES) (as_u16 [RES 3] [RES 2] [RES 1] [RES 0]))

;;
(defpurefun (as_u16 n3 n2 n1 n0) (+ (* 4096 n3) (* 256 n2) (* 16 n1) n0))
;; from stdlib
(defpurefun (or! (a :bool) (b :bool)) (if a (== 0 0) b))
(defpurefun (next X) (shift X 1))
(defpurefun (prev X) (shift X -1))
